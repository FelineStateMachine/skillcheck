package catalog

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed migrations/001_initial.sql
var initialMigration string

//go:embed migrations/002_skills_episodes.sql
var analysisMigration string

//go:embed migrations/003_workflows.sql
var workflowMigration string

type Catalog struct{ db *sql.DB }

func Open(path string) (*Catalog, error) {
	if path == "" {
		return nil, fmt.Errorf("catalog path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create catalog directory: %w", err)
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("protect catalog directory: %w", err)
	}
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open catalog: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(initialMigration + "\n" + analysisMigration + "\n" + workflowMigration); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate catalog: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		db.Close()
		return nil, fmt.Errorf("protect catalog: %w", err)
	}
	return &Catalog{db: db}, nil
}

func (c *Catalog) Close() error { return c.db.Close() }

func (c *Catalog) DB() *sql.DB { return c.db }

type Health struct {
	SourceKey      string `json:"source_key"`
	Harness        string `json:"harness"`
	Status         string `json:"status"`
	Revision       int64  `json:"revision"`
	EventCount     int    `json:"event_count"`
	ExclusionCount int    `json:"exclusion_count"`
}

func (c *Catalog) Health() ([]Health, error) {
	rows, err := c.db.Query(`SELECT source_key,harness,health,revision,event_count,exclusion_count FROM sources ORDER BY source_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Health
	for rows.Next() {
		var h Health
		if err := rows.Scan(&h.SourceKey, &h.Harness, &h.Status, &h.Revision, &h.EventCount, &h.ExclusionCount); err != nil {
			return nil, err
		}
		result = append(result, h)
	}
	return result, rows.Err()
}
