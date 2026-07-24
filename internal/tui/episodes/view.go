package episodes

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"skilltrace/internal/text"
	"skilltrace/internal/tui/viewport"
)

type Styles struct{ Title, Heading, Selected, Muted lipgloss.Style }

const (
	actorColumn   = 12
	outcomeColumn = 10
	tierColumn    = 10
	minRows       = 1
	// detailRows is what the expanded selection costs: capabilities and tokens.
	detailRows = 2
)

// View renders the episode list within the terminal height. Episodes of the
// same skill look alike, so each row carries its index and event range to make
// position in a long list legible.
func (m Model) View(styles Styles, unicode bool) string {
	width := m.Width
	if width <= 0 {
		width = 80
	}

	header := []string{
		styles.Title.Render(" SKILLTRACE / EPISODES "),
		styles.Heading.Render(m.heading()),
		"",
	}
	footer := []string{"", styles.Muted.Render(text.Clip(hints(unicode), width))}

	lines := make([]string, 0, len(header)+len(footer)+8)
	lines = append(lines, header...)
	lines = append(lines, m.rows(styles, unicode, width, m.listHeight(len(header), len(footer)))...)
	lines = append(lines, footer...)
	return strings.Join(lines, "\n")
}

func (m Model) heading() string {
	if len(m.Episodes) == 0 {
		return m.Skill
	}
	return fmt.Sprintf("%s (%d/%d)", m.Skill, m.Cursor+1, len(m.Episodes))
}

func (m Model) listHeight(header, footer int) int {
	if m.Height <= 0 {
		return len(m.Episodes)*(detailRows+1) + 1
	}
	return max(m.Height-header-footer, minRows)
}

func (m Model) rows(styles Styles, unicode bool, width, height int) []string {
	if len(m.Episodes) == 0 {
		return []string{styles.Muted.Render("No attributed uses in the cached sources.")}
	}

	marker := ">"
	if unicode {
		marker = "›"
	}
	// The selected row expands and an overflow indicator may follow the list,
	// so both come out of the budget before the window is sized.
	rowBudget := max(height-detailRows, minRows)
	if len(m.Episodes) > rowBudget {
		rowBudget = max(rowBudget-1, minRows)
	}
	start, end := viewport.Window(m.Cursor, len(m.Episodes), rowBudget)

	rows := make([]string, 0, height)
	for i := start; i < end; i++ {
		episode := m.Episodes[i]
		prefix := " "
		if i == m.Cursor {
			prefix = marker
		}
		line := fmt.Sprintf("%s %s %s %s %s evidence:%d",
			prefix,
			text.Cell(fmt.Sprintf("#%d", i+1), 7),
			text.Cell(fmt.Sprintf("seq %d-%d", episode.Start, episode.End), 16),
			text.Cell(episode.Actor, actorColumn),
			text.Cell(string(episode.Tier), tierColumn),
			len(episode.Evidence),
		)
		line = text.Clip(line, width)
		if i == m.Cursor {
			line = styles.Selected.Render(line)
		}
		rows = append(rows, line)
		if i == m.Cursor {
			rows = append(rows, styles.Muted.Render(text.Clip("    outcome: "+episode.Outcome+"   capabilities: "+strings.Join(episode.Capabilities, ", "), width)))
			if len(episode.Evidence) > 0 {
				rows = append(rows, styles.Muted.Render(text.Clip("    tokens: "+strings.Join(episode.Evidence, ", "), width)))
			}
		}
	}
	if hidden := viewport.Bar(start, end, len(m.Episodes)); hidden.Below > 0 {
		rows = append(rows, styles.Muted.Render(text.Clip(fmt.Sprintf("  ... %d more", hidden.Below), width)))
	}
	return rows
}

func hints(unicode bool) string {
	move := "up/down"
	if unicode {
		move = "↑/↓"
	}
	return move + " move  w workflow  esc back  q quit"
}
