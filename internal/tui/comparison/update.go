package comparison

import "skilltrace/internal/layout"

type MoveMsg int
type ResizeMsg struct{ Width, Height int }
type ToggleEvidenceMsg struct{}

func (m Model) Update(msg any) Model {
	switch v := msg.(type) {
	case MoveMsg:
		m.Difference += int(v)
		if m.Difference < 0 {
			m.Difference = 0
		}
		if last := len(m.Comparison.Differences) - 1; last >= 0 && m.Difference > last {
			m.Difference = last
		}
	case ResizeMsg:
		m.Width, m.Height = v.Width, v.Height
		m.Plan = layout.Compare(m.Comparison.Left.Graph, m.Comparison.Right.Graph, m.Comparison.Differences, v.Width)
	case ToggleEvidenceMsg:
		m.EvidenceOpen = !m.EvidenceOpen
	}
	return m
}
