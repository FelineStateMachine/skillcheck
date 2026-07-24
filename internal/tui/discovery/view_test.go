package discovery

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"skilltrace/internal/app"
)

func TestViewASCIIAndNoUsesStates(t *testing.T) {
	m := New(app.DiscoverySnapshot{
		Skills: []app.SkillSnapshot{{Name: "nzip", Description: "archive helper", State: "no uses"}},
		Sources: []app.SourceSnapshot{
			{Harness: "codex", Status: "supported", EventCount: 8, Freshness: "cached"},
			{Harness: "codex", Status: "supported", EventCount: 4, Freshness: "cached"},
		},
	})
	m.Width = 80
	plain := lipgloss.NewStyle()
	view := m.View(Styles{plain, plain, plain, plain, plain, plain}, false)
	// Sources are summarised per harness: two codex sessions, twelve events.
	for _, want := range []string{"SKILLTRACE / DISCOVERY", "> nzip", "no uses", "codex", "2 sessions", "12 events"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "›") {
		t.Fatal("ASCII view contains Unicode selection marker")
	}
}

func TestSourcesAreSummarisedPerHarnessNotListed(t *testing.T) {
	// Machine-wide sync yields thousands of sources; the view must not render
	// one row each or the skill list is buried.
	var sources []app.SourceSnapshot
	for range 500 {
		sources = append(sources, app.SourceSnapshot{Harness: "claude", Status: "supported", EventCount: 10, Freshness: "cached"})
	}
	m := New(app.DiscoverySnapshot{Skills: []app.SkillSnapshot{{Name: "nzip"}}, Sources: sources})
	m.Width, m.Height = 80, 24
	plain := lipgloss.NewStyle()
	view := m.View(Styles{plain, plain, plain, plain, plain, plain}, false)

	if lines := strings.Count(view, "\n") + 1; lines > 24 {
		t.Fatalf("view is %d lines for 500 sources; sources must be summarised", lines)
	}
	if !strings.Contains(view, "500 sessions") || !strings.Contains(view, "5000 events") {
		t.Fatalf("expected a per-harness summary line:\n%s", view)
	}
}
