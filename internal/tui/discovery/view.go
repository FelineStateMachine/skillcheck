package discovery

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"skilltrace/internal/text"
	"skilltrace/internal/tui/source"
	"skilltrace/internal/tui/viewport"
)

type Styles struct{ Title, Heading, Selected, Muted, Good, Warning lipgloss.Style }

const (
	nameColumn = 24
	// Wide enough for a skill exposed by both harnesses at one scope
	// ("claude+codex/global"); wider combinations clip.
	stateColumn = 20
	minRows     = 1
)

// View renders discovery into exactly the space the terminal has. The skill
// list is the only elastic region: chrome, sources, status and the footer are
// laid out first and the list takes whatever rows are left.
func (m Model) View(styles Styles, unicode bool) string {
	width := m.Width
	if width <= 0 {
		width = 80
	}

	header := m.header(styles)
	footer := m.footer(styles, unicode, width)
	rows := m.listRows(styles, unicode, width, m.listHeight(len(header), len(footer)))

	lines := make([]string, 0, len(header)+len(rows)+len(footer))
	lines = append(lines, header...)
	lines = append(lines, rows...)
	lines = append(lines, footer...)
	return strings.Join(lines, "\n")
}

func (m Model) header(styles Styles) []string {
	return []string{
		styles.Title.Render(" SKILLTRACE / DISCOVERY "),
		styles.Heading.Render(m.listHeading()),
	}
}

func (m Model) listHeading() string {
	total := len(m.Snapshot.Skills)
	if total == 0 {
		return "Installed skills"
	}
	return fmt.Sprintf("Installed skills (%d/%d)", m.Cursor+1, total)
}

// listHeight is what remains of the terminal after the fixed chrome. When the
// terminal is too short for even one row the list still gets one, so the
// selection never disappears entirely.
func (m Model) listHeight(header, footer int) int {
	if m.Height <= 0 {
		return len(m.Snapshot.Skills)
	}
	return max(m.Height-header-footer, minRows)
}

func (m Model) listRows(styles Styles, unicode bool, width, height int) []string {
	if len(m.Snapshot.Skills) == 0 {
		return m.emptyState(styles)
	}

	marker := ">"
	if unicode {
		marker = "›"
	}
	// Descriptions double the cost of every row, so they are only worth
	// showing when the list is not already fighting for space.
	withDescriptions := width >= 72 && height >= len(m.Snapshot.Skills)*2
	perRow := 1
	if withDescriptions {
		perRow = 2
	}

	// An overflow indicator follows the list when it does not all fit, so it
	// comes out of the budget before the window is sized.
	rowBudget := max(height/perRow, minRows)
	if len(m.Snapshot.Skills) > rowBudget {
		rowBudget = max(rowBudget-1, minRows)
	}
	start, end := viewport.Window(m.Cursor, len(m.Snapshot.Skills), rowBudget)
	rows := make([]string, 0, height)
	for i := start; i < end; i++ {
		skill := m.Snapshot.Skills[i]
		prefix := " "
		if i == m.Cursor {
			prefix = marker
		}
		line := prefix + " " + text.Cell(skill.Name, nameColumn) + " " + text.Clip(skill.State, stateColumn)
		line = text.Clip(line, width)
		if i == m.Cursor {
			line = styles.Selected.Render(line)
		}
		rows = append(rows, line)
		if withDescriptions && skill.Description != "" {
			rows = append(rows, styles.Muted.Render(text.Clip("    "+skill.Description, width)))
		}
	}
	if hidden := viewport.Bar(start, end, len(m.Snapshot.Skills)); hidden.Below > 0 {
		rows = append(rows, styles.Muted.Render(text.Clip(fmt.Sprintf("  ... %d more", hidden.Below), width)))
	}
	return rows
}

// emptyState names the directories that were searched. "No installed skills
// found" on its own gives the user nothing to act on.
func (m Model) emptyState(styles Styles) []string {
	if len(m.Roots) == 0 {
		return []string{styles.Muted.Render("  No installed skills found")}
	}
	lines := []string{styles.Muted.Render("  No installed skills found. Searched:")}
	for _, root := range m.Roots {
		lines = append(lines, styles.Muted.Render("    "+root))
	}
	return append(lines, styles.Muted.Render("  Override with --skill-root or SKILLTRACE_SKILL_ROOT."))
}

func (m Model) footer(styles Styles, unicode bool, width int) []string {
	lines := []string{"", styles.Heading.Render("Sources")}
	if len(m.Snapshot.Sources) == 0 {
		lines = append(lines, styles.Muted.Render("  No cached source snapshots"))
	}
	for _, snapshot := range m.Snapshot.Sources {
		row := source.Model{Harness: snapshot.Harness, Status: snapshot.Status, Revision: snapshot.Revision, Stale: snapshot.Freshness == "stale"}
		line := fmt.Sprintf("  %s %d events / %d excluded (rev %d, %s)",
			text.Cell(row.Label(), 28), snapshot.EventCount, snapshot.ExclusionCount, snapshot.Revision, snapshot.Freshness)
		style := styles.Good
		if row.Stale {
			style = styles.Warning
		}
		lines = append(lines, style.Render(text.Clip(line, width)))
	}

	// Discovery already records why a skill directory was skipped; dropping
	// those silently looks like the skills simply do not exist.
	if issues := m.Snapshot.Issues; len(issues) > 0 {
		reasons := make([]string, 0, len(issues))
		for _, issue := range issues {
			reasons = append(reasons, issue.Entry+" ("+issue.Reason+")")
		}
		summary := fmt.Sprintf("  %d skill(s) could not be read: %s", len(issues), strings.Join(reasons, ", "))
		lines = append(lines, styles.Warning.Render(text.Clip(summary, width)))
	}

	if m.Analyzing {
		lines = append(lines, "", styles.Warning.Render(text.Clip("Analyzing "+m.AnalyzingSkill+"... (esc to cancel)", width)))
	}
	if m.Scanning {
		lines = append(lines, "", styles.Warning.Render(text.Clip(m.progress(), width)))
	}
	if m.Error != "" {
		lines = append(lines, "", styles.Warning.Render(text.Clip("Scan failed: "+m.Error, width)))
	}

	return append(lines, "", styles.Muted.Render(text.Clip(m.hints(unicode), width)))
}

func (m Model) progress() string {
	label := "Scanning trace source: " + m.Stage
	if m.Total > 0 {
		return fmt.Sprintf("%s %d%%", label, min(m.Done*100/m.Total, 100))
	}
	return label
}

func (m Model) hints(unicode bool) string {
	move := "up/down"
	if unicode {
		move = "↑/↓"
	}
	hints := []string{move + " move", "enter episodes"}
	if m.ScanAvailable {
		hints = append(hints, "s scan")
	}
	return strings.Join(append(hints, "? help", "q quit"), "  ")
}
