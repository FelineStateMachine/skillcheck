package discovery

import (
	"testing"

	"skilltrace/internal/app"
)

func TestUpdateKeepsFocusInBoundsAndTracksProgress(t *testing.T) {
	m := New(app.DiscoverySnapshot{Skills: []app.SkillSnapshot{{Name: "a"}, {Name: "b"}}})
	m = m.Update(MoveMsg(4))
	if m.Cursor != 1 {
		t.Fatalf("cursor = %d", m.Cursor)
	}
	m = m.Update(ScanProgressMsg{Stage: "parsing", Done: 1, Total: 2})
	if !m.Scanning || m.Stage != "parsing" || m.Done != 1 {
		t.Fatalf("progress not retained: %#v", m)
	}
	m = m.Update(ResizeMsg{Width: 44, Height: 18})
	if m.Width != 44 || m.Height != 18 {
		t.Fatalf("resize not retained: %#v", m)
	}
}
