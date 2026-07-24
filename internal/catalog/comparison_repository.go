package catalog

import (
	"context"
	"encoding/json"
	"fmt"

	"skilltrace/internal/workflow"
)

func (c *Catalog) SaveComparison(ctx context.Context, comparison workflow.Comparison) error {
	b, err := json.Marshal(comparison)
	if err != nil {
		return fmt.Errorf("marshal comparison: %w", err)
	}
	_, err = c.db.ExecContext(ctx, `INSERT INTO cohort_comparisons(id, analyzer_revision, comparison_json) VALUES(1,?,?) ON CONFLICT(id) DO UPDATE SET analyzer_revision=excluded.analyzer_revision, comparison_json=excluded.comparison_json, updated_at=CURRENT_TIMESTAMP`, comparison.Revision, b)
	if err != nil {
		return fmt.Errorf("save comparison: %w", err)
	}
	return nil
}
