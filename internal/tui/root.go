package tui

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"skilltrace/internal/app"
	"skilltrace/internal/tui/discovery"
	"skilltrace/internal/tui/episodes"
)

type Route uint8

const (
	DiscoveryRoute Route = iota
	EpisodesRoute
)

type Config struct {
	SkillRoot string
	ScanInput string
	Harness   string
	Color     bool
	Unicode   bool
}

type Root struct {
	app       *app.Application
	config    Config
	Route     Route
	Discovery discovery.Model
	Episodes  episodes.Model
	theme     Theme
	Help      bool
	width     int
	height    int
	nextID    uint64
	activeID  uint64
	cancel    context.CancelFunc
}

type scanStreamMsg struct {
	ID       uint64
	Progress *app.Progress
	Result   *app.ScanResult
	Err      error
	stream   <-chan scanStreamMsg
}

func New(application *app.Application, config Config, snapshot app.DiscoverySnapshot) *Root {
	return &Root{app: application, config: config, Discovery: discovery.New(snapshot), theme: DefaultTheme(config.Color)}
}

func Load(application *app.Application, config Config) (*Root, error) {
	discovered, err := application.Discover(app.DiscoverRequest{SkillRoot: config.SkillRoot})
	if err != nil {
		return nil, err
	}
	health, err := application.Health()
	if err != nil {
		return nil, err
	}
	return New(application, config, app.PresentDiscovery(discovered, health, time.Now())), nil
}

func (m *Root) Init() tea.Cmd { return nil }

func (m *Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.Discovery = m.Discovery.Update(discovery.ResizeMsg{Width: msg.Width, Height: msg.Height})
		m.Episodes.Width, m.Episodes.Height = msg.Width, msg.Height
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	case OperationProgressMsg:
		if msg.ID == m.activeID {
			m.Discovery = m.Discovery.Update(discovery.ScanProgressMsg{Stage: msg.Stage, Done: msg.Completed, Total: msg.Total})
		}
	case OperationFinishedMsg:
		if msg.ID != m.activeID {
			return m, nil
		}
		m.cancel = nil
		switch result := msg.Result.(type) {
		case app.ScanResult:
			m.Discovery = m.Discovery.Update(discovery.ScanFinishedMsg{Err: msg.Err})
			if msg.Err == nil {
				return m, m.reloadCmd(msg.ID)
			}
		case app.DiscoverySnapshot:
			if msg.Err == nil {
				cursor := m.Discovery.Cursor
				m.Discovery = discovery.New(result)
				m.Discovery.Cursor = cursor
				m.Discovery = m.Discovery.Update(discovery.ResizeMsg{Width: m.width, Height: m.height})
			}
		case app.AnalyzeResult:
			if msg.Err == nil {
				m.Episodes = episodes.New(result.Skill, app.PresentEpisodes(result))
				m.Route = EpisodesRoute
			}
		}
	case scanStreamMsg:
		if msg.ID != m.activeID {
			return m, nil
		}
		if msg.Progress != nil {
			m.Discovery = m.Discovery.Update(discovery.ScanProgressMsg{Stage: msg.Progress.Stage, Done: msg.Progress.Completed, Total: msg.Progress.Total})
			return m, waitScan(msg.stream)
		}
		m.cancel = nil
		m.Discovery = m.Discovery.Update(discovery.ScanFinishedMsg{Err: msg.Err})
		if msg.Err == nil && msg.Result != nil {
			return m, m.reloadCmd(msg.ID)
		}
	}
	return m, nil
}

func (m *Root) updateKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if isKey(msg, "q", "ctrl+c") {
		m.stopOperation()
		return m, tea.Quit
	}
	if isKey(msg, "?") {
		m.Help = !m.Help
		return m, nil
	}
	if m.Help {
		if isKey(msg, "esc") {
			m.Help = false
		}
		return m, nil
	}
	if m.Route == EpisodesRoute {
		if isKey(msg, "esc", "left", "backspace") {
			m.Route = DiscoveryRoute
			return m, nil
		}
		if isKey(msg, "up", "k") && m.Episodes.Cursor > 0 {
			m.Episodes.Cursor--
		}
		if isKey(msg, "down", "j") && m.Episodes.Cursor+1 < len(m.Episodes.Episodes) {
			m.Episodes.Cursor++
		}
		return m, nil
	}
	if isKey(msg, "up", "k") {
		m.Discovery = m.Discovery.Update(discovery.MoveMsg(-1))
	}
	if isKey(msg, "down", "j") {
		m.Discovery = m.Discovery.Update(discovery.MoveMsg(1))
	}
	if isKey(msg, "enter") {
		if skill, ok := m.Discovery.Selected(); ok {
			id, ctx := m.beginOperation()
			return m, func() tea.Msg {
				result, err := m.app.Analyze(ctx, app.AnalyzeRequest{Skill: skill.Name, Scope: "current"})
				return OperationFinishedMsg{ID: id, Result: result, Err: err}
			}
		}
	}
	if isKey(msg, "s") && m.config.ScanInput != "" {
		id, ctx := m.beginOperation()
		m.Discovery = m.Discovery.Update(discovery.ScanStartedMsg{})
		stream := make(chan scanStreamMsg, 4)
		go func() {
			result, err := m.app.Scan(ctx, app.ScanRequest{Input: m.config.ScanInput, Harness: m.config.Harness}, func(progress app.Progress) {
				stream <- scanStreamMsg{ID: id, Progress: &progress, stream: stream}
			})
			stream <- scanStreamMsg{ID: id, Result: &result, Err: err, stream: stream}
			close(stream)
		}()
		return m, waitScan(stream)
	}
	if isKey(msg, "esc") && m.cancel != nil {
		m.stopOperation()
		m.Discovery = m.Discovery.Update(discovery.ScanFinishedMsg{Err: context.Canceled})
	}
	return m, nil
}

func waitScan(stream <-chan scanStreamMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-stream
		if !ok {
			return nil
		}
		return msg
	}
}

func (m *Root) beginOperation() (uint64, context.Context) {
	m.stopOperation()
	m.nextID++
	m.activeID = m.nextID
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	return m.activeID, ctx
}

func (m *Root) stopOperation() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
}

func (m *Root) reloadCmd(id uint64) tea.Cmd {
	return func() tea.Msg {
		discovered, err := m.app.Discover(app.DiscoverRequest{SkillRoot: m.config.SkillRoot})
		if err != nil {
			return OperationFinishedMsg{ID: id, Err: err}
		}
		health, err := m.app.Health()
		return OperationFinishedMsg{ID: id, Result: app.PresentDiscovery(discovered, health, time.Now()), Err: err}
	}
}

func (m *Root) View() tea.View {
	var content string
	if m.Route == EpisodesRoute {
		content = m.Episodes.View(episodes.Styles{Title: m.theme.Title, Heading: m.theme.Heading, Selected: m.theme.Selected, Muted: m.theme.Muted})
	} else {
		content = m.Discovery.View(discovery.Styles{Title: m.theme.Title, Heading: m.theme.Heading, Selected: m.theme.Selected, Muted: m.theme.Muted, Good: m.theme.Good, Warning: m.theme.Warning}, m.config.Unicode)
	}
	if m.Help {
		content += "\n\n" + m.theme.Heading.Render("Keyboard help") + "\n" + fmt.Sprintf("%s\n%s", "arrows/j/k: move   enter: open", "s: scan   esc: back/cancel   q: quit")
	}
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}
