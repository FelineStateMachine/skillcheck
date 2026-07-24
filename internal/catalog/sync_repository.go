package catalog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

// SyncMark records that a file of a given size and modification time has been
// scanned, keyed by a digest of its path so the path itself is never stored.
type SyncMark struct {
	Size         int64
	ModTimeNanos int64
}

func pathDigest(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:16])
}

// SyncMarks returns the recorded state of every previously synced file, keyed
// by path digest, so a sync run can decide what changed without re-reading.
func (c *Catalog) SyncMarks(ctx context.Context) (map[string]SyncMark, error) {
	rows, err := c.db.QueryContext(ctx, `SELECT path_digest, size, mod_time_nanos FROM sync_state`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	marks := map[string]SyncMark{}
	for rows.Next() {
		var digest string
		var mark SyncMark
		if err := rows.Scan(&digest, &mark.Size, &mark.ModTimeNanos); err != nil {
			return nil, err
		}
		marks[digest] = mark
	}
	return marks, rows.Err()
}

// RecordSync stores the state of a file after it has been scanned.
func (c *Catalog) RecordSync(ctx context.Context, path, harness string, mark SyncMark, scannedAt string) error {
	_, err := c.db.ExecContext(ctx,
		`INSERT INTO sync_state(path_digest,harness,size,mod_time_nanos,scanned_at) VALUES(?,?,?,?,?)
		 ON CONFLICT(path_digest) DO UPDATE SET harness=excluded.harness,size=excluded.size,mod_time_nanos=excluded.mod_time_nanos,scanned_at=excluded.scanned_at`,
		pathDigest(path), harness, mark.Size, mark.ModTimeNanos, scannedAt)
	return err
}

// SyncUnchanged reports whether a file matches its recorded size and mtime, so
// an unchanged file can be skipped without opening it.
func SyncUnchanged(marks map[string]SyncMark, path string, mark SyncMark) bool {
	previous, ok := marks[pathDigest(path)]
	return ok && previous == mark
}
