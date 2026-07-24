package discovery

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

type Styles struct{ Title, Heading, Selected, Muted, Good, Warning lipgloss.Style }

func (m Model) View(styles Styles, unicode bool) string {
	marker := ">"
	if unicode {
		marker = "›"
	}
	var b strings.Builder
	b.WriteString(styles.Title.Render(" SKILLTRACE / DISCOVERY "))
	b.WriteString("\n")
	b.WriteString(styles.Heading.Render("Installed skills"))
	b.WriteString("\n")
	if len(m.Snapshot.Skills) == 0 {
		b.WriteString(styles.Muted.Render("  No installed skills found"))
		b.WriteString("\n")
	}
	for i, skill := range m.Snapshot.Skills {
		prefix := " "
		line := fmt.Sprintf("%s %-20s %s", prefix, skill.Name, skill.State)
		if i == m.Cursor {
			line = fmt.Sprintf("%s %-20s %s", marker, skill.Name, skill.State)
			line = styles.Selected.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
		if m.Width >= 72 && skill.Description != "" {
			b.WriteString(styles.Muted.Render("    " + clip(skill.Description, m.Width-6)))
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(styles.Heading.Render("Sources"))
	b.WriteString("\n")
	if len(m.Snapshot.Sources) == 0 {
		b.WriteString(styles.Muted.Render("  No cached source snapshots"))
		b.WriteString("\n")
	}
	for _, source := range m.Snapshot.Sources {
		line := fmt.Sprintf("  %-10s %-12s %d events / %d excluded (%s)", source.Harness, source.Status, source.EventCount, source.ExclusionCount, source.Freshness)
		b.WriteString(styles.Good.Render(clip(line, m.Width)))
		b.WriteString("\n")
	}
	if m.Scanning {
		progress := fmt.Sprintf("Scanning current repository: %s %d/%d", m.Stage, m.Done, m.Total)
		b.WriteString("\n" + styles.Warning.Render(progress) + "\n")
	}
	if m.Error != "" {
		b.WriteString("\n" + styles.Warning.Render("Scan failed: "+m.Error) + "\n")
	}
	b.WriteString("\n" + styles.Muted.Render("↑/↓ move  enter episodes  s scan  ? help  q quit"))
	return b.String()
}

func clip(s string, width int) string {
	if width <= 0 || len([]rune(s)) <= width {
		return s
	}
	if width <= 3 {
		return string([]rune(s)[:width])
	}
	return string([]rune(s)[:width-3]) + "..."
}
