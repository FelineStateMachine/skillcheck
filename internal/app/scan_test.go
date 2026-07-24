package app_test

import (
	"context"
	"testing"

	"skilltrace/internal/app"
	"skilltrace/internal/catalog"
)

func TestScanCommitsTerminalSnapshot(t *testing.T) {
	c, err := catalog.Open(t.TempDir() + "/catalog.db")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	var progress []app.Progress
	result, err := app.New(c).Scan(context.Background(), app.ScanRequest{Input: "../../testdata/traces/codex/minimal.jsonl", Harness: "codex"}, func(p app.Progress) { progress = append(progress, p) })
	if err != nil {
		t.Fatal(err)
	}
	if result.Source.Revision != 1 || result.Source.EventCount != 6 || len(progress) != 2 || progress[1].Stage != "committed" {
		t.Fatalf("unexpected result: %+v progress=%+v", result, progress)
	}
}
