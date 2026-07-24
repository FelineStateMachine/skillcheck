CREATE TABLE IF NOT EXISTS workflow_passes (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  analyzer_revision TEXT NOT NULL,
  graph_json BLOB NOT NULL,
  metrics_json BLOB NOT NULL,
  findings_json BLOB NOT NULL,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
