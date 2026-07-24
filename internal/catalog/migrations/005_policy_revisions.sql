CREATE TABLE IF NOT EXISTS policy_revisions (
  revision TEXT PRIMARY KEY,
  fingerprint TEXT NOT NULL,
  applied_at TEXT NOT NULL,
  source_label TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS policy_applications (
  id INTEGER PRIMARY KEY,
  revision TEXT NOT NULL REFERENCES policy_revisions(revision),
  preview_token TEXT NOT NULL UNIQUE,
  affected_corrections INTEGER NOT NULL,
  affected_findings INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS candidates_policy_match ON candidates(skill_name, automated_label, actor);
