package layout

import (
	"skilltrace/internal/workflow"
	"testing"
)

func fixture() workflow.Graph {
	return workflow.Graph{Nodes: []workflow.Node{{ID: "invoke:skill", Stage: workflow.StageInvoke, Action: "skill"}, {ID: "act:tool", Stage: workflow.StageAct, Action: "tool"}}, Edges: []workflow.Edge{{From: "invoke:skill", To: "act:tool"}}}
}
func TestGoldenTerminalGeometry(t *testing.T) {
	p := Terminal(fixture(), 100)
	if p.Fallback || len(p.Nodes) != 2 || len(p.Routes) != 1 {
		t.Fatalf("unexpected plan: %#v", p)
	}
	if p.Nodes[0].Bounds.Contains(Point{X: -1, Y: -1}) {
		t.Fatal("invalid hit target")
	}
}
func TestNarrowSequenceFallback(t *testing.T) {
	p := Terminal(fixture(), 40)
	if !p.Fallback {
		t.Fatal("expected narrow fallback")
	}
	for _, n := range p.Nodes {
		if n.Bounds.X != 2 {
			t.Fatal("fallback nodes must use one lane")
		}
	}
}
