package workflow

import (
	"skilltrace/internal/layout"
	core "skilltrace/internal/workflow"
)

type Model struct {
	Skill           string
	Graph           core.Graph
	Metrics         core.Metrics
	Findings        []core.Finding
	Shapes          []core.Shape
	Plan            layout.Plan
	Cursor, Variant int
	Width, Height   int
	ASCII           bool
	EvidenceOpen    bool
}

func New(skill string, graph core.Graph, metrics core.Metrics, findings []core.Finding, shapes []core.Shape, width int) Model {
	return Model{Skill: skill, Graph: graph, Metrics: metrics, Findings: findings, Shapes: shapes, Width: width, Plan: layout.Terminal(graph, width)}
}

func (m Model) Selected() (core.Node, bool) {
	if m.Cursor < 0 || m.Cursor >= len(m.rankedNodes()) {
		return core.Node{}, false
	}
	return m.rankedNodes()[m.Cursor], true
}
