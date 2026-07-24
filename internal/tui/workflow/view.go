package workflow

import (
	"fmt"
	"strings"

	"skilltrace/internal/tui/grid"
)

func (m Model) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "WORKFLOW MAP  %d uses / %d exact variants\n", m.Metrics.Samples, m.Metrics.Variants)
	if m.Metrics.Limited() {
		b.WriteString("LIMITED EVIDENCE - comparative claims withheld\n")
	}
	if m.Plan.Fallback {
		b.WriteString("Observed sequence (narrow layout)\n")
		for i, n := range m.Graph.Nodes {
			marker := "  "
			if i == m.Cursor {
				marker = "> "
			}
			fmt.Fprintf(&b, "%s%s [%s] x%d\n", marker, n.Action, n.Stage, n.Count)
		}
	} else {
		b.WriteString(renderPlan(m))
	}
	if len(m.Findings) > 0 {
		b.WriteString("\nFINDINGS\n")
		for _, f := range m.Findings {
			fmt.Fprintf(&b, "- %s: %s\n", f.Title, f.Detail)
		}
	}
	if selected, ok := m.Selected(); ok {
		fmt.Fprintf(&b, "\nSELECTED %s | observed %d times", selected.Action, selected.Count)
	}
	if m.EvidenceOpen {
		b.WriteString("\nEVIDENCE references are resolved locally on demand")
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderPlan(m Model) string {
	g := grid.New(m.Plan.Width, m.Plan.Height)
	horizontal, vertical, corner := '─', '│', '└'
	if m.ASCII {
		horizontal, vertical, corner = '-', '|', '+'
	}
	for _, route := range m.Plan.Routes {
		for i := 1; i < len(route.Points); i++ {
			a, z := route.Points[i-1], route.Points[i]
			if a.Y == z.Y {
				for x := min(a.X, z.X); x <= max(a.X, z.X); x++ {
					g.Set(x, a.Y, horizontal)
				}
			} else {
				for y := min(a.Y, z.Y); y <= max(a.Y, z.Y); y++ {
					g.Set(a.X, y, vertical)
				}
				g.Set(a.X, z.Y, corner)
			}
		}
	}
	for i, n := range m.Plan.Nodes {
		prefix, suffix := "[", "]"
		if i == m.Cursor {
			prefix, suffix = "{", "}"
		}
		g.Text(n.Bounds.X, n.Bounds.Y, prefix+n.Label+suffix)
	}
	return g.String() + "\n"
}
