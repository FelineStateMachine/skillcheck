package tui

import (
	"context"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"skilltrace/internal/app"
	"skilltrace/internal/text"
	"skilltrace/internal/tui/comparison"
	"skilltrace/internal/tui/discovery"
	"skilltrace/internal/tui/episodes"
	"skilltrace/internal/tui/policy"
	"skilltrace/internal/tui/workflow"
)

type Route uint8

const (
	DiscoveryRoute Route = iota
	EpisodesRoute
	WorkflowRoute
	ComparisonRoute
	PolicyRoute
)

type Config struct {
	SkillRoot  string
	SkillRoots []string
	ScanInput  string
	Harness    string
	Color      bool
	Unicode    bool
	// PolicyFile, CohortLeft and CohortRight mirror the inputs the headless
	// policy and compare commands take, so the same work is reachable from
	// the TUI without inventing a second way to specify them.
	PolicyFile  string
	CohortLeft  string
	CohortRight string
	Skill       string
}

type Root struct {
	app        *app.Application
	config     Config
	Route      Route
	Discovery  discovery.Model
	Episodes   episodes.Model
	Workflow   workflow.Model
	Comparison comparison.Model
	Policy     policy.Model
	theme      Theme
	Help       bool
	width      int
	height     int
	nextID     uint64
	activeID   uint64
	cancel     context.CancelFunc
	// analysis retains the last analyze so the workflow route can render
	// without recomputing what analyze already produced.
	analysis app.AnalyzeResult
	// status carries route-level failures that are not scan failures.
	status string
	// Interrupted records that the session ended on ctrl+c rather than q.
	Interrupted bool
}

type scanStreamMsg struct {
	ID       uint64
	Progress *app.Progress
	Result   *app.ScanResult
	Err      error
	stream   <-chan scanStreamMsg
}

func New(application *app.Application, config Config, snapshot app.DiscoverySnapshot) *Root {
	root := &Root{app: application, config: config, Discovery: discovery.New(snapshot), theme: DefaultTheme(config.Color)}
	root.Discovery.Roots = config.SkillRoots
	root.Discovery.ScanAvailable = config.ScanInput != ""
	root.Policy = policy.New(config.PolicyFile)
	return root
}

func Load(application *app.Application, config Config) (*Root, error) {
	request := app.DiscoverRequest{SkillRoot: config.SkillRoot}
	discovered, err := application.Discover(request)
	if err != nil {
		return nil, err
	}
	health, err := application.Health()
	if err != nil {
		return nil, err
	}
	config.SkillRoots = request.Roots()
	return New(application, config, app.PresentDiscovery(discovered, health, time.Now())), nil
}

func (m *Root) Init() tea.Cmd { return nil }

func (m *Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.Discovery = m.Discovery.Update(discovery.ResizeMsg{Width: msg.Width, Height: msg.Height})
		m.Episodes.Width, m.Episodes.Height = msg.Width, msg.Height
		m.Workflow = m.Workflow.Update(workflow.ResizeMsg{Width: msg.Width, Height: msg.Height})
		m.Comparison = m.Comparison.Update(comparison.ResizeMsg{Width: msg.Width, Height: msg.Height})
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	case OperationProgressMsg:
		if msg.ID == m.activeID {
			m.Discovery = m.Discovery.Update(discovery.ScanProgressMsg{Stage: msg.Stage, Done: msg.Completed, Total: msg.Total})
		}
	case OperationFinishedMsg:
		return m.finish(msg)
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

// finish routes a completed operation. Every branch has to handle msg.Err:
// swallowing it leaves a failed operation looking exactly like a slow one.
func (m *Root) finish(msg OperationFinishedMsg) (tea.Model, tea.Cmd) {
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
		if msg.Err != nil {
			m.Discovery = m.Discovery.Update(discovery.AnalyzeFinishedMsg{Err: msg.Err})
			return m, nil
		}
		cursor := m.Discovery.Cursor
		roots, available := m.Discovery.Roots, m.Discovery.ScanAvailable
		m.Discovery = discovery.New(result)
		m.Discovery.Cursor, m.Discovery.Roots, m.Discovery.ScanAvailable = cursor, roots, available
		m.Discovery = m.Discovery.Update(discovery.ResizeMsg{Width: m.width, Height: m.height})
	case app.AnalyzeResult:
		m.Discovery = m.Discovery.Update(discovery.AnalyzeFinishedMsg{Err: msg.Err})
		if msg.Err != nil {
			return m, nil
		}
		m.analysis = result
		m.Episodes = episodes.New(result.Skill, app.PresentEpisodes(result))
		m.Episodes.Width, m.Episodes.Height = m.width, m.height
		m.Route = EpisodesRoute
	case app.CompareResult:
		if msg.Err != nil {
			m.status = msg.Err.Error()
			return m, nil
		}
		m.Comparison = comparison.New(result.Comparison, m.width)
		m.Comparison.Height = m.height
		m.Route = ComparisonRoute
	case policyResult:
		m.Policy = result.apply(m.Policy, msg.Err)
		m.Route = PolicyRoute
	default:
		if msg.Err != nil {
			m.status = msg.Err.Error()
		}
	}
	return m, nil
}

func (m *Root) updateKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if isKey(msg, "q", "ctrl+c") {
		m.stopOperation()
		// README documents 130 for cancelled work; an interrupt should honour
		// that rather than reporting the same clean exit as pressing q.
		m.Interrupted = isKey(msg, "ctrl+c")
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
	switch m.Route {
	case EpisodesRoute:
		return m.updateEpisodesKey(msg)
	case WorkflowRoute:
		return m.updateWorkflowKey(msg)
	case ComparisonRoute:
		return m.updateComparisonKey(msg)
	case PolicyRoute:
		return m.updatePolicyKey(msg)
	}
	return m.updateDiscoveryKey(msg)
}

func (m *Root) updateEpisodesKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case isKey(msg, "esc", "left", "backspace"):
		m.Route = DiscoveryRoute
	case isKey(msg, "up", "k") && m.Episodes.Cursor > 0:
		m.Episodes.Cursor--
	case isKey(msg, "down", "j") && m.Episodes.Cursor+1 < len(m.Episodes.Episodes):
		m.Episodes.Cursor++
	case isKey(msg, "w"):
		m.Workflow = workflow.New(m.analysis.Skill, m.analysis.Workflow.Graph, m.analysis.Workflow.Metrics, m.analysis.Workflow.Findings, m.analysis.Workflow.Shapes, m.width)
		m.Workflow.Height = m.height
		m.Workflow.ASCII = !m.config.Unicode
		m.Route = WorkflowRoute
	}
	return m, nil
}

func (m *Root) updateWorkflowKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case isKey(msg, "esc", "left", "backspace"):
		m.Route = EpisodesRoute
	case isKey(msg, "up", "k"):
		m.Workflow = m.Workflow.Update(workflow.MoveMsg(-1))
	case isKey(msg, "down", "j"):
		m.Workflow = m.Workflow.Update(workflow.MoveMsg(1))
	case isKey(msg, "e"):
		m.Workflow = m.Workflow.Update(workflow.ToggleEvidenceMsg{})
	}
	return m, nil
}

func (m *Root) updateComparisonKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case isKey(msg, "esc", "left", "backspace"):
		m.Route = DiscoveryRoute
	case isKey(msg, "up", "k"):
		m.Comparison = m.Comparison.Update(comparison.MoveMsg(-1))
	case isKey(msg, "down", "j"):
		m.Comparison = m.Comparison.Update(comparison.MoveMsg(1))
	case isKey(msg, "e"):
		m.Comparison = m.Comparison.Update(comparison.ToggleEvidenceMsg{})
	}
	return m, nil
}

func (m *Root) updatePolicyKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case isKey(msg, "esc", "left", "backspace"):
		m.Route = DiscoveryRoute
	case isKey(msg, "v"):
		return m, m.policyCmd(policyValidate)
	case isKey(msg, "enter"):
		return m, m.policyCmd(policyPreview)
	case isKey(msg, "a"):
		// Applying rewrites catalog rows, so it is gated on a preview token
		// bound to the current revision rather than fired on a keypress alone.
		gate := m.applyGate()
		if gate.Confirmable(m.Policy.Preview.Token) {
			return m, m.policyCmd(policyApply)
		}
		if gate.WriterBusy {
			m.Policy = m.Policy.Fail("another catalog operation is in flight")
			return m, nil
		}
		m.Policy = m.Policy.Fail("preview the policy before applying it")
	}
	return m, nil
}

func (m *Root) updateDiscoveryKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if isKey(msg, "esc") && m.cancel != nil {
		m.stopOperation()
		m.Discovery = m.Discovery.Update(discovery.ScanFinishedMsg{Err: context.Canceled})
		m.Discovery = m.Discovery.Update(discovery.AnalyzeFinishedMsg{})
		return m, nil
	}
	if isKey(msg, "up", "k") {
		m.Discovery = m.Discovery.Update(discovery.MoveMsg(-1))
	}
	if isKey(msg, "down", "j") {
		m.Discovery = m.Discovery.Update(discovery.MoveMsg(1))
	}
	// Repeat activations while busy used to cancel and restart the work, so
	// mashing enter through a slow analyze made it strictly slower.
	if m.Discovery.Busy() {
		return m, nil
	}
	if isKey(msg, "enter") {
		if skill, ok := m.Discovery.Selected(); ok {
			id, ctx := m.beginOperation()
			m.Discovery = m.Discovery.Update(discovery.AnalyzeStartedMsg{Skill: skill.Name})
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
	if isKey(msg, "p") && m.config.PolicyFile != "" {
		m.Route = PolicyRoute
		return m, nil
	}
	if isKey(msg, "c") && m.config.CohortLeft != "" && m.config.CohortRight != "" {
		skill := m.config.Skill
		if selected, ok := m.Discovery.Selected(); ok && skill == "" {
			skill = selected.Name
		}
		id, ctx := m.beginOperation()
		return m, func() tea.Msg {
			result, err := m.app.Compare(ctx, app.CompareRequest{Skill: skill, LeftPath: m.config.CohortLeft, RightPath: m.config.CohortRight})
			return OperationFinishedMsg{ID: id, Result: result, Err: err}
		}
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
	width := m.width
	if width <= 0 {
		width = 80
	}
	var content string
	switch m.Route {
	case EpisodesRoute:
		content = m.Episodes.View(episodes.Styles{Title: m.theme.Title, Heading: m.theme.Heading, Selected: m.theme.Selected, Muted: m.theme.Muted}, m.config.Unicode)
	case WorkflowRoute:
		content = m.Workflow.View(workflow.Styles{Title: m.theme.Title, Heading: m.theme.Heading, Selected: m.theme.Selected, Muted: m.theme.Muted, Good: m.theme.Good, Warning: m.theme.Warning})
	case ComparisonRoute:
		content = m.Comparison.View()
	case PolicyRoute:
		content = m.Policy.View()
	default:
		content = m.Discovery.View(discovery.Styles{Title: m.theme.Title, Heading: m.theme.Heading, Selected: m.theme.Selected, Muted: m.theme.Muted, Good: m.theme.Good, Warning: m.theme.Warning}, m.config.Unicode)
	}
	if m.status != "" {
		content += "\n" + m.theme.Warning.Render(text.Clip(m.status, width))
	}
	if m.Help {
		content = m.helpView(width)
	}
	view := tea.NewView(m.fit(content))
	view.AltScreen = true
	return view
}

// helpView replaces the screen instead of appending to it. Appending put the
// help text below the fold on any list long enough to need help.
func (m *Root) helpView(width int) string {
	lines := []string{
		m.theme.Title.Render(" SKILLTRACE / HELP "),
		"",
		m.theme.Heading.Render("Keyboard help"),
		"  up/down, j/k   move",
		"  enter          open episodes for the selected skill",
		"  w              workflow map (from episodes)",
		"  e              toggle evidence (workflow, comparison)",
		"  esc            back, or cancel a running operation",
		"  q, ctrl+c      quit",
	}
	if m.config.ScanInput != "" {
		lines = append(lines, "  s              scan the configured trace source")
	}
	if m.config.PolicyFile != "" {
		lines = append(lines, "  p              policy: v validate, enter preview, a apply")
	}
	if m.config.CohortLeft != "" && m.config.CohortRight != "" {
		lines = append(lines, "  c              compare the configured cohorts")
	}
	lines = append(lines, "", m.theme.Muted.Render("? or esc to close"))
	for i, line := range lines {
		lines[i] = text.Clip(line, width)
	}
	return strings.Join(lines, "\n")
}

// fit trims the rendered view to the terminal height. Views size themselves,
// but this guarantees no route can push the footer off the bottom.
func (m *Root) fit(content string) string {
	if m.height <= 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	if len(lines) <= m.height {
		return content
	}
	return strings.Join(lines[:m.height], "\n")
}
