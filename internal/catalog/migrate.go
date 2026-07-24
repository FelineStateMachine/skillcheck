package catalog

import (
	"context"
	"database/sql"
	"fmt"
)

type migration struct {
	version int
	sql     string
}

func migrate(ctx context.Context, db *sql.DB) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY)`); err != nil {
		return err
	}
	migrations := []migration{{1, initialMigration}, {2, analysisMigration}, {3, workflowMigration}, {4, cohortMigration}, {5, policyMigration}, {6, lifecycleMigration}}
	for _, m := range migrations {
		var found int
		scanErr := tx.QueryRowContext(ctx, `SELECT version FROM schema_migrations WHERE version=?`, m.version).Scan(&found)
		if scanErr == nil {
			continue
		}
		if scanErr != sql.ErrNoRows {
			return scanErr
		}
		if _, err = tx.ExecContext(ctx, m.sql); err != nil {
			return fmt.Errorf("migration %d: %w", m.version, err)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES(?)`, m.version); err != nil {
			return err
		}
	}
	return tx.Commit()
}
