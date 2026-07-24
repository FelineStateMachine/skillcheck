package catalog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type DeletionPreview struct {
	Action   string `json:"action"`
	Target   string `json:"target"`
	Token    string `json:"token"`
	Revision int64  `json:"revision"`
	Sources  int    `json:"sources"`
	Skills   int    `json:"skills"`
	Episodes int    `json:"episodes"`
}

func (c *Catalog) Preview(ctx context.Context, action, target string) (DeletionPreview, error) {
	p := DeletionPreview{Action: action, Target: target}
	var err error
	switch action {
	case "forget-source":
		err = c.db.QueryRowContext(ctx, `SELECT revision,1 FROM sources WHERE source_key=?`, target).Scan(&p.Revision, &p.Sources)
	case "clear-skill":
		err = c.db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(MAX(s.revision),0) FROM candidates c JOIN sources s ON s.id=c.source_id WHERE c.skill_name=?`, target).Scan(&p.Episodes, &p.Revision)
		p.Skills = 1
	case "reset":
		err = c.db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(MAX(revision),0) FROM sources`).Scan(&p.Sources, &p.Revision)
		_ = c.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM skills`).Scan(&p.Skills)
		_ = c.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM episodes`).Scan(&p.Episodes)
	default:
		return p, fmt.Errorf("unknown lifecycle action")
	}
	if err != nil {
		return p, err
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%d\x00%d\x00%d\x00%d", action, target, p.Revision, p.Sources, p.Skills, p.Episodes)))
	p.Token = hex.EncodeToString(sum[:16])
	return p, nil
}

func (c *Catalog) Execute(ctx context.Context, preview DeletionPreview) (err error) {
	current, err := c.Preview(ctx, preview.Action, preview.Target)
	if err != nil {
		return err
	}
	if current.Token != preview.Token {
		return fmt.Errorf("deletion preview is stale")
	}
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	switch preview.Action {
	case "forget-source":
		_, err = tx.ExecContext(ctx, `DELETE FROM sources WHERE source_key=?`, preview.Target)
	case "clear-skill":
		_, err = tx.ExecContext(ctx, `DELETE FROM candidates WHERE skill_name=?`, preview.Target)
		if err == nil {
			_, err = tx.ExecContext(ctx, `DELETE FROM skills WHERE name=?`, preview.Target)
		}
	case "reset":
		for _, table := range []string{"dependency_state", "source_checkpoints", "policy_applications", "policy_revisions", "workflow_variants", "cohort_memberships", "skills", "sources"} {
			if _, e := tx.ExecContext(ctx, `DELETE FROM `+table); e != nil { /* optional phase tables may differ */
			}
		}
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}
