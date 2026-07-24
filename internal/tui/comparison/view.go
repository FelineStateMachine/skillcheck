package comparison

import (
	"fmt"
	"strings"
)

func (m Model) View() string {
	c := m.Comparison
	var b strings.Builder
	fmt.Fprintf(&b, "COHORT COMPARISON  %s vs %s\n", c.Left.Definition.Label, c.Right.Definition.Label)
	fmt.Fprintf(&b, "%-24s | %s\n", fmt.Sprintf("LEFT  %d uses", len(c.Left.Members)), fmt.Sprintf("RIGHT %d uses", len(c.Right.Members)))
	for _, warning := range c.Left.Warnings {
		fmt.Fprintf(&b, "LEFT WARNING: %s\n", warning)
	}
	for _, warning := range c.Right.Warnings {
		fmt.Fprintf(&b, "RIGHT WARNING: %s\n", warning)
	}
	if !c.Compatibility.Usage {
		b.WriteString("UNSUPPORTED: usage metric is not available on both sides\n")
	}
	b.WriteString("\nDIFFERENCES\n")
	for i, d := range c.Differences {
		marker := "  "
		if i == m.Difference {
			marker = "> "
		}
		fmt.Fprintf(&b, "%s%s [%s] %d | %d - %s\n", marker, d.Action, d.Stage, d.LeftCount, d.RightCount, d.Detail)
	}
	if d, ok := m.Selected(); ok {
		fmt.Fprintf(&b, "\nSYNCHRONIZED FOCUS: %s on both complete workflows", d.ID)
	}
	if m.EvidenceOpen {
		b.WriteString("\nEVIDENCE references are resolved locally on demand")
	}
	return strings.TrimRight(b.String(), "\n")
}
