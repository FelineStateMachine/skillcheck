package headless

import (
	"fmt"
	"skilltrace/internal/app"
)

func ExportText(result app.ExportResult) string {
	return fmt.Sprintf("exported %s report to %s", result.Format, result.Output)
}
