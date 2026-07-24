package workflow

import (
	"skilltrace/internal/layout"
	core "skilltrace/internal/workflow"
)

type Model struct {
	Graph           core.Graph
	Metrics         core.Metrics
	Findings        []core.Finding
	Plan            layout.Plan
	Cursor, Variant int
	Width, Height   int
	ASCII           bool
	EvidenceOpen    bool
}

func New(graph core.Graph, metrics core.Metrics, findings []core.Finding, width int) Model {
	return Model{Graph: graph, Metrics: metrics, Findings: findings, Width: width, Plan: layout.Terminal(graph, width)}
}
func (m Model) Selected() (core.Node, bool) {
	if m.Cursor < 0 || m.Cursor >= len(m.Graph.Nodes) {
		return core.Node{}, false
	}
	return m.Graph.Nodes[m.Cursor], true
}
