package workflow

import "skilltrace/internal/layout"

type MoveMsg int
type ResizeMsg struct{ Width, Height int }
type NextVariantMsg struct{}
type ToggleEvidenceMsg struct{}

func (m Model) Update(msg any) Model {
	switch v := msg.(type) {
	case MoveMsg:
		m.Cursor += int(v)
		if m.Cursor < 0 {
			m.Cursor = 0
		}
		if last := len(m.Graph.Nodes) - 1; last >= 0 && m.Cursor > last {
			m.Cursor = last
		}
	case ResizeMsg:
		m.Width, m.Height = v.Width, v.Height
		m.Plan = layout.Terminal(m.Graph, v.Width)
	case NextVariantMsg:
		if len(m.Graph.Variants) > 0 {
			m.Variant = (m.Variant + 1) % len(m.Graph.Variants)
		}
	case ToggleEvidenceMsg:
		m.EvidenceOpen = !m.EvidenceOpen
	}
	return m
}
