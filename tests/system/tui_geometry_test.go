package system_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"skilltrace/internal/app"
	"skilltrace/internal/catalog"
	"skilltrace/internal/tui"
	"skilltrace/internal/tui/episodes"
)

// The discovery and episode views must stay inside the terminal the user
// actually has. Overflowing the height hides the selection marker, the source
// inventory, and the footer with no scroll affordance to recover them.

func geometryRoot(t *testing.T, snapshot app.DiscoverySnapshot, width, height int) *tui.Root {
	t.Helper()
	c, err := catalog.Open(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	root := tui.New(app.New(c), tui.Config{}, snapshot)
	_, _ = root.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return root
}

func snapshotWithSkills(n int) app.DiscoverySnapshot {
	snapshot := app.DiscoverySnapshot{
		Sources: []app.SourceSnapshot{{Harness: "codex", Status: "supported", EventCount: 12, Freshness: "cached"}},
	}
	for i := range n {
		snapshot.Skills = append(snapshot.Skills, app.SkillSnapshot{
			Name:        fmt.Sprintf("skill-%02d", i),
			Description: strings.Repeat("description ", 8),
			State:       "cached",
		})
	}
	return snapshot
}

func press(root *tui.Root, key tea.Key, times int) {
	for range times {
		_, _ = root.Update(tea.KeyPressMsg(key))
	}
}

func lines(view string) []string { return strings.Split(view, "\n") }

func TestDiscoveryViewFitsTerminalHeight(t *testing.T) {
	const height = 24
	root := geometryRoot(t, snapshotWithSkills(26), 80, height)
	got := len(lines(root.View().Content))
	if got > height {
		t.Fatalf("discovery view rendered %d lines into a %d-row terminal; everything past the fold is unreachable", got, height)
	}
}

func TestDiscoveryKeepsCursorAndFooterVisibleWhileScrolling(t *testing.T) {
	const height = 24
	root := geometryRoot(t, snapshotWithSkills(26), 80, height)
	press(root, tea.Key{Code: tea.KeyDown}, 20)

	view := root.View().Content
	if !strings.Contains(view, "> skill-20") {
		t.Fatalf("cursor moved to skill-20 but no selection marker is on screen:\n%s", view)
	}
	if !strings.Contains(view, "q quit") {
		t.Fatalf("footer scrolled off screen:\n%s", view)
	}
	if !strings.Contains(view, "Sources") {
		t.Fatalf("source inventory scrolled off screen:\n%s", view)
	}
	if got := len(lines(view)); got > height {
		t.Fatalf("view is %d lines in a %d-row terminal", got, height)
	}
}

func TestDiscoveryHelpIsVisibleWhenToggled(t *testing.T) {
	const height = 24
	root := geometryRoot(t, snapshotWithSkills(26), 80, height)
	_, _ = root.Update(tea.KeyPressMsg(tea.Key{Text: "?", Code: '?'}))

	view := root.View().Content
	if !strings.Contains(view, "Keyboard help") {
		t.Fatalf("help was toggled on but is not on screen:\n%s", view)
	}
	if got := len(lines(view)); got > height {
		t.Fatalf("help view is %d lines in a %d-row terminal", got, height)
	}
}

func TestEpisodesViewFitsTerminalHeightAndKeepsCursorVisible(t *testing.T) {
	const height = 24
	eps := make([]app.EpisodeSnapshot, 500)
	for i := range eps {
		eps[i] = app.EpisodeSnapshot{
			Actor: "primary", Outcome: "unknown", Tier: "confirmed",
			Capabilities: []string{"actor", "outcome", "evidence"},
			Evidence:     []string{"event-2", "event-3"},
		}
	}
	root := geometryRoot(t, snapshotWithSkills(1), 80, height)
	root.Route = tui.EpisodesRoute
	root.Episodes = episodes.New("nzip", eps)
	_, _ = root.Update(tea.WindowSizeMsg{Width: 80, Height: height})
	press(root, tea.Key{Code: tea.KeyDown}, 40)

	view := root.View().Content
	if got := len(lines(view)); got > height {
		t.Fatalf("episodes view rendered %d lines into a %d-row terminal", got, height)
	}
	if !strings.Contains(view, ">") {
		t.Fatalf("no selection marker on screen after scrolling:\n%s", view)
	}
	if !strings.Contains(view, "q quit") {
		t.Fatalf("footer scrolled off screen:\n%s", view)
	}
}

func TestNarrowTerminalDoesNotOverflowWidth(t *testing.T) {
	const width = 40
	root := geometryRoot(t, snapshotWithSkills(6), width, 20)
	for _, line := range lines(root.View().Content) {
		if len([]rune(line)) > width {
			t.Fatalf("line exceeds %d columns (%d): %q", width, len([]rune(line)), line)
		}
	}
}
