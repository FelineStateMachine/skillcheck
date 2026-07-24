package catalog

import (
	"context"
	"encoding/json"
	"skilltrace/internal/detection"
	"skilltrace/internal/trace"
)

func (c *Catalog) CurrentEvents(ctx context.Context) ([]trace.Event, error) {
	rows, err := c.db.QueryContext(ctx, `SELECT e.sequence,e.kind,e.contract_version,e.coordinate_json,e.payload_json FROM events e JOIN snapshots s ON s.id=e.snapshot_id ORDER BY e.sequence`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []trace.Event
	for rows.Next() {
		var e trace.Event
		var coordinate, payload string
		if err := rows.Scan(&e.Sequence, &e.Kind, &e.Version, &coordinate, &payload); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(coordinate), &e.Coordinate); err != nil {
			return nil, err
		}
		e.Payload = []byte(payload)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (c *Catalog) SaveEpisodes(ctx context.Context, skill string, episodes []detection.Episode) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM candidates WHERE skill_name=?`, skill); err != nil {
		return err
	}
	for _, e := range episodes {
		r, err := tx.ExecContext(ctx, `INSERT INTO candidates(source_id,skill_name,automated_label,final_label,actor,start_sequence,end_sequence,analyzer_revision) SELECT id,?,?,?,?,?,?,? FROM sources ORDER BY updated_at DESC LIMIT 1`, skill, e.Tier, e.Tier, e.Actor, e.Start, e.End, "detection-v1")
		if err != nil {
			return err
		}
		cid, _ := r.LastInsertId()
		r, err = tx.ExecContext(ctx, `INSERT INTO episodes(candidate_id,outcome) VALUES(?,?)`, cid, e.Outcome)
		if err != nil {
			return err
		}
		eid, _ := r.LastInsertId()
		for _, seq := range e.EventSequences {
			if _, err = tx.ExecContext(ctx, `INSERT INTO episode_events(episode_id,sequence) VALUES(?,?)`, eid, seq); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}
