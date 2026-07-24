package headless

import (
	"fmt"
	"strings"

	"skilltrace/internal/app"
)

func CompareText(result app.CompareResult) string {
	c := result.Comparison
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %s (%d uses) vs %s (%d uses)\n", result.Skill, c.Left.Definition.Label, len(c.Left.Members), c.Right.Definition.Label, len(c.Right.Members))
	if c.ClaimsWithheld {
		b.WriteString("LIMITED EVIDENCE - comparative claims withheld\n")
	}
	for _, d := range c.Differences {
		fmt.Fprintf(&b, "- %s: %d vs %d (%s)\n", d.Action, d.LeftCount, d.RightCount, d.Detail)
	}
	return strings.TrimSpace(b.String())
}
