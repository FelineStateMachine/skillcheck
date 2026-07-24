CREATE TABLE IF NOT EXISTS cohort_comparisons (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  analyzer_revision TEXT NOT NULL,
  comparison_json BLOB NOT NULL,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
