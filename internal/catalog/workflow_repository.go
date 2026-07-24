package catalog

import (
	"context"
	"encoding/json"
	"skilltrace/internal/workflow"
)

func (c *Catalog) SaveWorkflow(ctx context.Context, graph workflow.Graph, metrics workflow.Metrics, findings []workflow.Finding) error {
	g, _ := json.Marshal(graph)
	m, _ := json.Marshal(metrics)
	f, _ := json.Marshal(findings)
	_, err := c.db.ExecContext(ctx, `INSERT INTO workflow_passes(id,analyzer_revision,graph_json,metrics_json,findings_json) VALUES(1,'workflow-v1',?,?,?) ON CONFLICT(id) DO UPDATE SET analyzer_revision=excluded.analyzer_revision,graph_json=excluded.graph_json,metrics_json=excluded.metrics_json,findings_json=excluded.findings_json,updated_at=CURRENT_TIMESTAMP`, g, m, f)
	return err
}
