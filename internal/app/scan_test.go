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
	if result.Source.Revision != 1 || result.Source.EventCount != 6 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(progress) < 2 || progress[0].Stage != "parsing" || progress[len(progress)-1].Stage != "committed" {
		t.Fatalf("scan must report parsing then committed, got %+v", progress)
	}
}

// Progress has to track the source actually being consumed; a fixed 0/1 then
// 1/1 tells the user nothing about a large trace.
func TestScanReportsParseProgressAgainstSourceSize(t *testing.T) {
	c, err := catalog.Open(t.TempDir() + "/catalog.db")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	var progress []app.Progress
	_, err = app.New(c).Scan(context.Background(), app.ScanRequest{Input: "../../testdata/traces/codex/minimal.jsonl", Harness: "codex"}, func(p app.Progress) { progress = append(progress, p) })
	if err != nil {
		t.Fatal(err)
	}
	var advanced bool
	for _, p := range progress {
		if p.Stage == "parsing" && p.Total > 0 && p.Completed > 0 {
			advanced = true
		}
		if p.Total > 0 && p.Completed > p.Total {
			t.Fatalf("progress overshot: %+v", p)
		}
	}
	if !advanced {
		t.Fatalf("no parse progress was reported: %+v", progress)
	}
}
