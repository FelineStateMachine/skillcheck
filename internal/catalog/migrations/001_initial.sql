CREATE TABLE IF NOT EXISTS sources (
    id INTEGER PRIMARY KEY,
    source_key TEXT NOT NULL UNIQUE,
    harness TEXT NOT NULL,
    health TEXT NOT NULL,
    revision INTEGER NOT NULL DEFAULT 0,
    event_count INTEGER NOT NULL DEFAULT 0,
    exclusion_count INTEGER NOT NULL DEFAULT 0,
    capabilities_json TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS snapshots (
    id INTEGER PRIMARY KEY,
    source_id INTEGER NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    revision INTEGER NOT NULL,
    committed_at TEXT NOT NULL,
    UNIQUE(source_id, revision)
);
CREATE TABLE IF NOT EXISTS events (
    snapshot_id INTEGER NOT NULL REFERENCES snapshots(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL,
    kind TEXT NOT NULL,
    contract_version INTEGER NOT NULL,
    coordinate_json TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    PRIMARY KEY(snapshot_id, sequence)
);
CREATE TABLE IF NOT EXISTS exclusions (
    snapshot_id INTEGER NOT NULL REFERENCES snapshots(id) ON DELETE CASCADE,
    line INTEGER NOT NULL,
    reason TEXT NOT NULL
);
