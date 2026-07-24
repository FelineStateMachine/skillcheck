package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"skilltrace/internal/adapters"
)

type Snapshot struct {
	Revision                   int64
	EventCount, ExclusionCount int
}

func (c *Catalog) ReplaceSource(ctx context.Context, sourceKey string, result adapters.Result) (snapshot Snapshot, err error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return snapshot, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	capabilities, err := json.Marshal(result.Capabilities)
	if err != nil {
		return snapshot, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = tx.ExecContext(ctx, `INSERT INTO sources(source_key,harness,health,revision,event_count,exclusion_count,capabilities_json,updated_at) VALUES(?,?,?,1,?,?,?,?) ON CONFLICT(source_key) DO UPDATE SET harness=excluded.harness,health=excluded.health,revision=sources.revision+1,event_count=excluded.event_count,exclusion_count=excluded.exclusion_count,capabilities_json=excluded.capabilities_json,updated_at=excluded.updated_at`, sourceKey, result.Harness, "supported", len(result.Events), len(result.Exclusions), string(capabilities), now)
	if err != nil {
		return snapshot, fmt.Errorf("upsert source: %w", err)
	}
	var sourceID, revision int64
	if err = tx.QueryRowContext(ctx, `SELECT id,revision FROM sources WHERE source_key=?`, sourceKey).Scan(&sourceID, &revision); err != nil {
		return snapshot, err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM snapshots WHERE source_id=?`, sourceID); err != nil {
		return snapshot, err
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO snapshots(source_id,revision,committed_at) VALUES(?,?,?)`, sourceID, revision, now)
	if err != nil {
		return snapshot, err
	}
	snapshotID, err := res.LastInsertId()
	if err != nil {
		return snapshot, err
	}
	for _, event := range result.Events {
		coordinate, marshalErr := json.Marshal(event.Coordinate)
		if marshalErr != nil {
			return snapshot, marshalErr
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO events(snapshot_id,sequence,kind,contract_version,coordinate_json,payload_json) VALUES(?,?,?,?,?,?)`, snapshotID, event.Sequence, event.Kind, event.Version, string(coordinate), string(event.Payload)); err != nil {
			return snapshot, fmt.Errorf("insert event: %w", err)
		}
	}
	for _, exclusion := range result.Exclusions {
		if _, err = tx.ExecContext(ctx, `INSERT INTO exclusions(snapshot_id,line,reason) VALUES(?,?,?)`, snapshotID, exclusion.Line, exclusion.Reason); err != nil {
			return snapshot, err
		}
	}
	if err = tx.Commit(); err != nil {
		return snapshot, err
	}
	return Snapshot{Revision: revision, EventCount: len(result.Events), ExclusionCount: len(result.Exclusions)}, nil
}
