package workflow

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"

	"skilltrace/internal/text"
	core "skilltrace/internal/workflow"
)

type Styles struct{ Title, Heading, Selected, Muted, Good, Warning lipgloss.Style }

// View renders the workflow as a terminal-native dataviz: the common
// class-level paths a skill's uses followed, then every tool ranked by volume
// with its failure rate. This is the graph the old three-box diagram could not
// show, because node identity used to be the bare kind "tool".
func (m Model) View(styles Styles) string {
	width := m.Width
	if width <= 0 {
		width = 80
	}
	var b strings.Builder

	title := fmt.Sprintf(" WORKFLOW · %s · %d uses · %d variants ", m.Skill, m.Metrics.Samples, m.Metrics.Variants)
	b.WriteString(styles.Title.Render(text.Clip(title, width)))
	b.WriteString("\n")
	if m.Metrics.Limited() {
		b.WriteString(styles.Muted.Render(text.Clip("limited evidence — comparative claims withheld", width)))
		b.WriteString("\n")
	}

	m.writeShapes(&b, styles, width)
	m.writeTools(&b, styles, width)
	m.writeFindings(&b, styles, width)

	b.WriteString("\n" + styles.Muted.Render(text.Clip("up/down tool  esc back  q quit", width)))
	return b.String()
}

func (m Model) writeShapes(b *strings.Builder, styles Styles, width int) {
	if len(m.Shapes) == 0 {
		return
	}
	b.WriteString("\n" + styles.Heading.Render("Common paths") + "\n")
	top := m.Shapes[0].Count
	for _, shape := range m.Shapes {
		bar := miniBar(shape.Count, top, 10)
		line := fmt.Sprintf("  %4d  %s  %s", shape.Count, bar, shape.String())
		b.WriteString(styles.Good.Render(text.Clip(line, width)) + "\n")
	}
}

func (m Model) writeTools(b *strings.Builder, styles Styles, width int) {
	nodes := m.rankedNodes()
	if len(nodes) == 0 {
		return
	}
	b.WriteString("\n" + styles.Heading.Render("Tools by volume") + "\n")
	maxCount := nodes[0].Count
	// The bar competes with fixed columns for width; give it what is left.
	barWidth := max(width-40, 8)
	for i, node := range nodes {
		marker := " "
		if i == m.Cursor {
			marker = ">"
		}
		bar := miniBar(node.Count, maxCount, barWidth)
		errRate := ""
		if node.Errors > 0 && node.Count > 0 {
			errRate = fmt.Sprintf(" %d%% err", node.Errors*100/node.Count)
		}
		line := fmt.Sprintf("%s %s %s %s %d%s",
			marker,
			text.Cell(node.Action, 16),
			text.Cell(string(node.Class), 8),
			bar, node.Count, errRate)
		rendered := text.Clip(line, width)
		switch {
		case i == m.Cursor:
			rendered = styles.Selected.Render(rendered)
		case node.Errors > 0:
			rendered = styles.Warning.Render(rendered)
		}
		b.WriteString(rendered + "\n")
	}
}

func (m Model) writeFindings(b *strings.Builder, styles Styles, width int) {
	if len(m.Findings) == 0 {
		return
	}
	b.WriteString("\n" + styles.Heading.Render("Findings") + "\n")
	for _, f := range m.Findings {
		b.WriteString(styles.Muted.Render(text.Clip("  - "+f.Title+": "+f.Detail, width)) + "\n")
	}
}

// rankedNodes orders the graph's nodes by volume, so the busiest tools lead.
func (m Model) rankedNodes() []core.Node {
	nodes := append([]core.Node(nil), m.Graph.Nodes...)
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Count != nodes[j].Count {
			return nodes[i].Count > nodes[j].Count
		}
		return nodes[i].Action < nodes[j].Action
	})
	return nodes
}

func miniBar(value, top, width int) string {
	if top <= 0 || width <= 0 {
		return ""
	}
	filled := value * width / top
	if filled < 1 && value > 0 {
		filled = 1
	}
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("·", width-filled)
}
