package comparison

import (
	"skilltrace/internal/layout"
	core "skilltrace/internal/workflow"
)

type Model struct {
	Comparison    core.Comparison
	Plan          layout.ComparisonPlan
	Width, Height int
	Difference    int
	EvidenceOpen  bool
}

func New(value core.Comparison, width int) Model {
	return Model{Comparison: value, Width: width, Plan: layout.Compare(value.Left.Graph, value.Right.Graph, value.Differences, width)}
}

func (m Model) Selected() (core.Difference, bool) {
	if m.Difference < 0 || m.Difference >= len(m.Comparison.Differences) {
		return core.Difference{}, false
	}
	return m.Comparison.Differences[m.Difference], true
}
