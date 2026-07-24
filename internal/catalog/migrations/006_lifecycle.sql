CREATE TABLE IF NOT EXISTS catalog_metadata (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
INSERT OR IGNORE INTO catalog_metadata(key,value) VALUES('schema_version','6');
INSERT OR IGNORE INTO catalog_metadata(key,value) VALUES('minimum_reader_version','1');
CREATE TABLE IF NOT EXISTS source_checkpoints (
    source_id INTEGER PRIMARY KEY REFERENCES sources(id) ON DELETE CASCADE,
    committed_offset INTEGER NOT NULL,
    parser_revision TEXT NOT NULL,
    fingerprint_json TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS dependency_state (
    product_kind TEXT NOT NULL,
    product_key TEXT NOT NULL,
    source_revision INTEGER NOT NULL,
    policy_revision TEXT NOT NULL DEFAULT '',
    stale INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY(product_kind,product_key)
);
CREATE TABLE IF NOT EXISTS writer_lease (
    singleton INTEGER PRIMARY KEY CHECK(singleton=1),
    owner TEXT NOT NULL,
    expires_at INTEGER NOT NULL
);
