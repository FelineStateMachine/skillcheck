package comparison

import (
	"strings"
	"testing"

	core "skilltrace/internal/workflow"
)

func TestViewShowsCompleteCohortsAndSynchronizedFocus(t *testing.T) {
	c := core.Comparison{Left: core.Cohort{Definition: core.CohortDefinition{Label: "Codex"}, Warnings: []string{"limited evidence"}}, Right: core.Cohort{Definition: core.CohortDefinition{Label: "Claude"}}, Differences: []core.Difference{{ID: "execute:tool", Stage: "execute", Action: "tool", Detail: "visible", LeftCount: 2, RightCount: 1}}}
	view := New(c, 100).View()
	for _, want := range []string{"Codex vs Claude", "LEFT WARNING", "SYNCHRONIZED FOCUS", "tool"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q: %s", want, view)
		}
	}
}
