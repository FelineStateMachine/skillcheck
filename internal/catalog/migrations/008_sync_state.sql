-- Machine-wide sync walks thousands of trace files on every run. Re-parsing
-- all of them each time is wasteful, so a cheap size+mtime check per file lets
-- an incremental sync touch only what changed.
--
-- The key is a digest of the absolute path, not the path itself: docs/privacy.md
-- keeps source paths out of the catalog, and the filesystem walk rediscovers
-- the real paths on each run.
CREATE TABLE IF NOT EXISTS sync_state (
    path_digest    TEXT PRIMARY KEY,
    harness        TEXT NOT NULL,
    size           INTEGER NOT NULL,
    mod_time_nanos INTEGER NOT NULL,
    scanned_at     TEXT NOT NULL
);

UPDATE catalog_metadata SET value='8' WHERE key='schema_version';
