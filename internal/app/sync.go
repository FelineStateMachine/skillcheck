package app

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"skilltrace/internal/catalog"
)

// SyncRoot is a directory tree of traces that all belong to one harness.
type SyncRoot struct {
	Harness string
	Path    string
}

// SyncRequest selects what to walk. Empty Roots means the default machine-wide
// locations for every known harness.
type SyncRequest struct {
	Roots []SyncRoot
	Now   time.Time
}

// SyncResult reports what a run touched.
type SyncResult struct {
	Scanned int      `json:"scanned"`
	Skipped int      `json:"skipped"`
	Failed  int      `json:"failed"`
	Events  int      `json:"events"`
	Roots   []string `json:"roots"`
}

// SyncProgress is called as files are processed, so a long walk can show
// progress without the caller knowing the file count up front.
type SyncProgress func(scanned, skipped, failed int)

// DefaultSyncRoots returns the machine-wide trace locations for each harness.
func DefaultSyncRoots(home string) []SyncRoot {
	if home == "" {
		return nil
	}
	return []SyncRoot{
		{Harness: "claude", Path: filepath.Join(home, ".claude", "projects")},
		{Harness: "codex", Path: filepath.Join(home, ".codex", "sessions")},
	}
}

// Sync walks the requested roots and scans every trace file that is new or has
// changed since the last run, skipping the rest on a cheap size and mtime
// check. It is the command that turns an empty catalog into a picture of the
// whole machine.
//
// A file that fails to parse is counted and skipped rather than aborting the
// run: one truncated session must not stop the sync. Files that scan cleanly
// have their state recorded so the next run leaves them alone.
func (a *Application) Sync(ctx context.Context, req SyncRequest, progress SyncProgress) (SyncResult, error) {
	roots := req.Roots
	if len(roots) == 0 {
		home, _ := os.UserHomeDir()
		roots = DefaultSyncRoots(home)
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}
	scannedAt := now.UTC().Format(time.RFC3339Nano)

	marks, err := a.Catalog.SyncMarks(ctx)
	if err != nil {
		return SyncResult{}, err
	}

	var result SyncResult
	for _, root := range roots {
		result.Roots = append(result.Roots, root.Path)
		err := filepath.WalkDir(root.Path, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				// A missing root is not an error: a machine may run only one
				// harness. Unreadable subtrees are skipped.
				return nil //nolint:nilerr
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if d.IsDir() || filepath.Ext(path) != ".jsonl" {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			mark := catalog.SyncMark{Size: info.Size(), ModTimeNanos: info.ModTime().UnixNano()}
			if catalog.SyncUnchanged(marks, path, mark) {
				result.Skipped++
				a.report(progress, result)
				return nil
			}
			scan, scanErr := a.Scan(ctx, ScanRequest{Input: path, Harness: root.Harness}, nil)
			if scanErr != nil {
				result.Failed++
				a.report(progress, result)
				return nil
			}
			if err := a.Catalog.RecordSync(ctx, path, root.Harness, mark, scannedAt); err != nil {
				return err
			}
			result.Scanned++
			result.Events += scan.Source.EventCount
			a.report(progress, result)
			return nil
		})
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (a *Application) report(progress SyncProgress, r SyncResult) {
	if progress != nil {
		progress(r.Scanned, r.Skipped, r.Failed)
	}
}
