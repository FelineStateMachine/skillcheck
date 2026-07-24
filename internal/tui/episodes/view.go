package episodes

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

type Styles struct{ Title, Heading, Selected, Muted lipgloss.Style }

func (m Model) View(styles Styles) string {
	var b strings.Builder
	b.WriteString(styles.Title.Render(" SKILLTRACE / EPISODES "))
	b.WriteString("\n")
	b.WriteString(styles.Heading.Render(m.Skill))
	b.WriteString("\n\n")
	if len(m.Episodes) == 0 {
		b.WriteString(styles.Muted.Render("No attributed uses in the cached sources."))
		b.WriteString("\n")
	}
	for i, episode := range m.Episodes {
		line := fmt.Sprintf("  %-12s %-10s %-10s evidence:%d", episode.Actor, episode.Outcome, episode.Tier, len(episode.Evidence))
		if i == m.Cursor {
			line = styles.Selected.Render(">" + line[1:])
		}
		b.WriteString(line + "\n")
		if i == m.Cursor {
			b.WriteString(styles.Muted.Render("    capabilities: "+strings.Join(episode.Capabilities, ", ")) + "\n")
			if len(episode.Evidence) > 0 {
				b.WriteString(styles.Muted.Render("    tokens: "+strings.Join(episode.Evidence, ", ")) + "\n")
			}
		}
	}
	b.WriteString("\n" + styles.Muted.Render("↑/↓ move  esc back  q quit"))
	return b.String()
}
