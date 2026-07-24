package workflow

import (
	core "skilltrace/internal/workflow"
	"strings"
	"testing"
)

func graph() core.Graph {
	return core.Graph{Nodes: []core.Node{{ID: "invoke:skill", Stage: core.StageInvoke, Action: "skill", Count: 3}, {ID: "act:tool", Stage: core.StageAct, Action: "tool", Count: 2}}, Edges: []core.Edge{{From: "invoke:skill", To: "act:tool"}}}
}
func TestGoldenWorkflowView(t *testing.T) {
	got := New(graph(), core.Metrics{Samples: 3, Variants: 2}, nil, 100).View()
	if !strings.Contains(got, "WORKFLOW MAP") || !strings.Contains(got, "LIMITED EVIDENCE") {
		t.Fatalf("unexpected view:\n%s", got)
	}
}
func TestNarrowWorkflowView(t *testing.T) {
	got := New(graph(), core.Metrics{Samples: 3, Variants: 2}, nil, 40).View()
	if !strings.Contains(got, "Observed sequence (narrow layout)") {
		t.Fatalf("unexpected view:\n%s", got)
	}
}
func TestASCIIWorkflowView(t *testing.T) {
	m := New(graph(), core.Metrics{Samples: 6, Variants: 1}, nil, 100)
	m.ASCII = true
	got := m.View()
	if strings.ContainsAny(got, "─│└") {
		t.Fatalf("unicode connector in ASCII view:\n%s", got)
	}
}
