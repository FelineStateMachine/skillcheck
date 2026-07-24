package catalog

import (
	"database/sql"
	"fmt"
	"time"
)

type PolicyAudit struct {
	Revision            string `json:"revision"`
	PreviewToken        string `json:"preview_token"`
	AffectedCorrections int    `json:"affected_corrections"`
	AffectedFindings    int    `json:"affected_findings"`
}

func (c *Catalog) CountPolicyCandidates(clause string, args ...any) (int, error) {
	var count int
	err := c.db.QueryRow("SELECT COUNT(*) FROM candidates WHERE "+clause, args...).Scan(&count)
	return count, err
}

func (c *Catalog) ApplyPolicy(a PolicyAudit, fingerprint, source string) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`INSERT OR IGNORE INTO policy_revisions(revision,fingerprint,applied_at,source_label) VALUES(?,?,?,?)`, a.Revision, fingerprint, time.Now().UTC().Format(time.RFC3339), source); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO policy_applications(revision,preview_token,affected_corrections,affected_findings) VALUES(?,?,?,?)`, a.Revision, a.PreviewToken, a.AffectedCorrections, a.AffectedFindings); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("policy application failed")
		}
		return err
	}
	return tx.Commit()
}
