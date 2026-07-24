-- Event sequences are numbered per source, so analysis has to know which
-- source an event came from. Without that, windows and sequence-keyed lookups
-- silently mix unrelated sessions together.
--
-- Sessions also carry the metadata that makes machine-wide views possible:
-- the project a session ran in, its branch, and when it happened. Project is
-- stored as a bare directory name, never a full path (see docs/privacy.md).
CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY,
    source_id INTEGER NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    session_key TEXT NOT NULL,
    harness TEXT NOT NULL,
    project TEXT NOT NULL DEFAULT '',
    branch TEXT NOT NULL DEFAULT '',
    started_at TEXT NOT NULL DEFAULT '',
    ended_at TEXT NOT NULL DEFAULT '',
    UNIQUE(source_id, session_key)
);

CREATE INDEX IF NOT EXISTS sessions_project ON sessions(project);
CREATE INDEX IF NOT EXISTS events_snapshot_sequence ON events(snapshot_id, sequence);

UPDATE catalog_metadata SET value='7' WHERE key='schema_version';
