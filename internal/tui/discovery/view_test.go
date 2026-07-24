package discovery

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"skilltrace/internal/app"
)

func TestViewASCIIAndNoUsesStates(t *testing.T) {
	m := New(app.DiscoverySnapshot{
		Skills:  []app.SkillSnapshot{{Name: "nzip", Description: "archive helper", State: "no uses"}},
		Sources: []app.SourceSnapshot{{Harness: "codex", Status: "supported", EventCount: 8, Freshness: "cached"}},
	})
	m.Width = 80
	plain := lipgloss.NewStyle()
	view := m.View(Styles{plain, plain, plain, plain, plain, plain}, false)
	for _, want := range []string{"SKILLTRACE / DISCOVERY", "> nzip", "no uses", "codex", "8 events", "cached"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "›") {
		t.Fatal("ASCII view contains Unicode selection marker")
	}
}
