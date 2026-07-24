package workflow

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	core "skilltrace/internal/workflow"
)

func graph() core.Graph {
	return core.Graph{Nodes: []core.Node{
		{ID: "act:Bash", Stage: core.StageAct, Action: "Bash", Class: core.ClassExecute, Count: 40, Errors: 4},
		{ID: "act:Read", Stage: core.StageAct, Action: "Read", Class: core.ClassInspect, Count: 12},
		{ID: "invoke:skill", Stage: core.StageInvoke, Action: "skill", Class: core.ClassSkill, Count: 3},
	}}
}

func plainStyles() Styles {
	p := lipgloss.NewStyle()
	return Styles{p, p, p, p, p, p}
}

func TestViewRanksToolsByVolume(t *testing.T) {
	m := New("code-review", graph(), core.Metrics{Samples: 6, Variants: 2},
		nil, []core.Shape{{Classes: []core.Class{core.ClassSkill, core.ClassExecute}, Count: 6}}, 100)
	got := m.View(plainStyles())

	if !strings.Contains(got, "code-review") || !strings.Contains(got, "Tools by volume") {
		t.Fatalf("view missing header/tools:\n%s", got)
	}
	// Bash has the highest count, so it must appear before Read.
	if strings.Index(got, "Bash") > strings.Index(got, "Read") {
		t.Fatalf("tools not ranked by volume:\n%s", got)
	}
	// Real tool names and classes are shown, not the bare kind "tool".
	for _, want := range []string{"Bash", "execute", "Read", "inspect", "Common paths"} {
		if !strings.Contains(got, want) {
			t.Fatalf("view missing %q:\n%s", want, got)
		}
	}
}

func TestViewShowsErrorRate(t *testing.T) {
	m := New("code-review", graph(), core.Metrics{Samples: 6, Variants: 1}, nil, nil, 100)
	got := m.View(plainStyles())
	if !strings.Contains(got, "10% err") {
		t.Fatalf("Bash has 4 errors of 40 calls; expected 10%% err:\n%s", got)
	}
}

func TestViewFlagsLimitedEvidence(t *testing.T) {
	m := New("nzip", graph(), core.Metrics{Samples: 3, Variants: 2}, nil, nil, 100)
	if got := m.View(plainStyles()); !strings.Contains(got, "limited evidence") {
		t.Fatalf("expected limited-evidence note:\n%s", got)
	}
}

func TestViewFitsNarrowWidth(t *testing.T) {
	m := New("nzip", graph(), core.Metrics{Samples: 6, Variants: 1}, nil, nil, 40)
	for _, line := range strings.Split(m.View(plainStyles()), "\n") {
		if len([]rune(line)) > 40 {
			t.Fatalf("line exceeds 40 columns (%d): %q", len([]rune(line)), line)
		}
	}
}
