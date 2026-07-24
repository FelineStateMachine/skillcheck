package tui

import (
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"skilltrace/internal/app"
	"skilltrace/internal/catalog"
)

func TestRootResizeOverlayAndLateResult(t *testing.T) {
	c, err := catalog.Open(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	m := New(app.New(c), Config{}, app.DiscoverySnapshot{Skills: []app.SkillSnapshot{{Name: "nzip"}}})
	_, _ = m.Update(tea.WindowSizeMsg{Width: 41, Height: 13})
	if m.Discovery.Width != 41 || m.Discovery.Height != 13 {
		t.Fatal("resize was not routed")
	}
	m.activeID = 2
	_, _ = m.Update(OperationFinishedMsg{ID: 1, Result: app.DiscoverySnapshot{}})
	if len(m.Discovery.Snapshot.Skills) != 1 {
		t.Fatal("late result replaced current discovery")
	}
	_, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "?", Code: '?'}))
	if !m.Help {
		t.Fatal("help overlay did not open")
	}
}
