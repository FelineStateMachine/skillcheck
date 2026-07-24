package catalog

import (
	"context"
	"encoding/json"
	"skilltrace/internal/detection"
	"skilltrace/internal/trace"
)

// Session identifies one recorded run of a harness. Project is a bare
// directory name rather than a path, per docs/privacy.md.
type Session struct {
	Key       string `json:"key"`
	Harness   string `json:"harness"`
	Project   string `json:"project,omitempty"`
	Branch    string `json:"branch,omitempty"`
	StartedAt string `json:"started_at,omitempty"`
}

// SessionEvents is one session's events, in sequence order.
type SessionEvents struct {
	Session Session
	Events  []trace.Event
}

// CurrentSessions returns the committed events grouped by the source that
// produced them.
//
// Sequences are numbered per source, so a query that ordered globally by
// sequence interleaved unrelated sessions: episode windows absorbed another
// session's events, and sequence-keyed lookups downstream collided outright.
// Grouping here is what makes every sequence-keyed consumer safe.
func (c *Catalog) CurrentSessions(ctx context.Context) ([]SessionEvents, error) {
	rows, err := c.db.QueryContext(ctx, `
		SELECT s.id, src.source_key, src.harness,
		       COALESCE(ses.project,''), COALESCE(ses.branch,''), COALESCE(ses.started_at,''),
		       e.sequence, e.kind, e.contract_version, e.coordinate_json, e.payload_json
		FROM events e
		JOIN snapshots s ON s.id = e.snapshot_id
		JOIN sources src ON src.id = s.source_id
		LEFT JOIN sessions ses ON ses.source_id = src.id
		ORDER BY s.id, e.sequence`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SessionEvents
	var currentID int64 = -1
	for rows.Next() {
		var (
			snapshotID                        int64
			key, harness, project, branch, at string
			e                                 trace.Event
			coordinate, payload               string
		)
		if err := rows.Scan(&snapshotID, &key, &harness, &project, &branch, &at,
			&e.Sequence, &e.Kind, &e.Version, &coordinate, &payload); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(coordinate), &e.Coordinate); err != nil {
			return nil, err
		}
		e.Payload = []byte(payload)
		if snapshotID != currentID {
			currentID = snapshotID
			out = append(out, SessionEvents{Session: Session{Key: key, Harness: harness, Project: project, Branch: branch, StartedAt: at}})
		}
		last := &out[len(out)-1]
		last.Events = append(last.Events, e)
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
