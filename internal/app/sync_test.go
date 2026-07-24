package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"skilltrace/internal/app"
	"skilltrace/internal/catalog"
)

func writeTrace(t *testing.T, path string) {
	t.Helper()
	body := `{"type":"session","model":"gpt-5-codex"}` + "\n" +
		`{"type":"skill","skill":"demo"}` + "\n" +
		`{"type":"tool","tool":"shell","status":"completed"}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func syncApp(t *testing.T) *app.Application {
	t.Helper()
	c, err := catalog.Open(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return app.New(c)
}

func TestSyncWalksARootAndScansEveryTrace(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a.jsonl", "b.jsonl"} {
		writeTrace(t, filepath.Join(root, name))
	}
	// A nested directory must be reached too, and non-jsonl ignored.
	sub := filepath.Join(root, "2026", "07")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTrace(t, filepath.Join(sub, "c.jsonl"))
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("ignore me"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := syncApp(t).Sync(context.Background(), app.SyncRequest{
		Roots: []app.SyncRoot{{Harness: "codex", Path: root}},
		Now:   time.Unix(1, 0),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Scanned != 3 || result.Failed != 0 {
		t.Fatalf("expected 3 traces scanned, got %+v", result)
	}
	if result.Events == 0 {
		t.Fatal("sync reported no events across three traces")
	}
}

func TestSyncSkipsUnchangedFilesOnTheSecondRun(t *testing.T) {
	root := t.TempDir()
	writeTrace(t, filepath.Join(root, "a.jsonl"))
	writeTrace(t, filepath.Join(root, "b.jsonl"))
	a := syncApp(t)
	req := app.SyncRequest{Roots: []app.SyncRoot{{Harness: "codex", Path: root}}, Now: time.Unix(1, 0)}

	first, err := a.Sync(context.Background(), req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if first.Scanned != 2 || first.Skipped != 0 {
		t.Fatalf("first run should scan both: %+v", first)
	}

	// Nothing changed on disk, so a second run must scan nothing.
	second, err := a.Sync(context.Background(), req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if second.Scanned != 0 || second.Skipped != 2 {
		t.Fatalf("second run should skip both unchanged files: %+v", second)
	}
}

func TestSyncRescansAChangedFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a.jsonl")
	writeTrace(t, path)
	a := syncApp(t)
	req := app.SyncRequest{Roots: []app.SyncRoot{{Harness: "codex", Path: root}}, Now: time.Unix(1, 0)}

	if _, err := a.Sync(context.Background(), req, nil); err != nil {
		t.Fatal(err)
	}

	// Append a record and bump mtime; the file must be picked up again.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"type":"completion","status":"success"}` + "\n"); err != nil {
		t.Fatal(err)
	}
	f.Close()
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}

	second, err := a.Sync(context.Background(), req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if second.Scanned != 1 || second.Skipped != 0 {
		t.Fatalf("a changed file should be rescanned: %+v", second)
	}
}

func TestSyncMissingRootIsNotAnError(t *testing.T) {
	result, err := syncApp(t).Sync(context.Background(), app.SyncRequest{
		Roots: []app.SyncRoot{{Harness: "claude", Path: filepath.Join(t.TempDir(), "does-not-exist")}},
	}, nil)
	if err != nil {
		t.Fatalf("a machine without one harness should not fail sync: %v", err)
	}
	if result.Scanned != 0 {
		t.Fatalf("nothing to scan under a missing root: %+v", result)
	}
}
