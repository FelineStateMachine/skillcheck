package app_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"skilltrace/internal/app"
	"skilltrace/internal/catalog"
)

// Event sequences are numbered per source. Analysing a catalog that holds more
// than one source must therefore never let one source's events fall inside
// another source's episode windows, and must never let two sources collide on
// the same sequence number.
func TestAnalyzeIsolatesSourcesFromEachOther(t *testing.T) {
	trace := func(t *testing.T, dir, name, skill string, tools int) string {
		t.Helper()
		path := filepath.Join(dir, name)
		lines := []string{`{"type":"session","model":"gpt-5-codex"}`, `{"type":"skill","skill":"` + skill + `"}`}
		for i := range tools {
			lines = append(lines, `{"type":"tool","tool":"shell","status":"completed","payload":{"command":"echo `+string(rune('a'+i))+`"}}`)
		}
		lines = append(lines, `{"type":"completion","status":"success"}`)
		body := ""
		for _, l := range lines {
			body += l + "\n"
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}

	dir := t.TempDir()
	alpha := trace(t, dir, "alpha.jsonl", "alpha-skill", 2)
	beta := trace(t, dir, "beta.jsonl", "beta-skill", 5)

	analyze := func(t *testing.T, inputs []string, skill string) app.AnalyzeResult {
		t.Helper()
		c, err := catalog.Open(filepath.Join(t.TempDir(), "catalog.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		a := app.New(c)
		for _, in := range inputs {
			if _, err := a.Scan(context.Background(), app.ScanRequest{Input: in, Harness: "codex"}, nil); err != nil {
				t.Fatal(err)
			}
		}
		result, err := a.Analyze(context.Background(), app.AnalyzeRequest{Skill: skill, Scope: "current"})
		if err != nil {
			t.Fatal(err)
		}
		return result
	}

	for _, tc := range []struct {
		skill string
		only  []string
	}{
		{"alpha-skill", []string{alpha}},
		{"beta-skill", []string{beta}},
	} {
		alone := analyze(t, tc.only, tc.skill)
		together := analyze(t, []string{alpha, beta}, tc.skill)

		wantJSON, _ := json.Marshal(alone.Episodes)
		gotJSON, _ := json.Marshal(together.Episodes)
		if string(wantJSON) != string(gotJSON) {
			t.Fatalf("%s: episodes changed when a second source was present in the catalog\n alone: %s\n  both: %s",
				tc.skill, wantJSON, gotJSON)
		}
	}
}

// The workflow projection keys events by sequence. Two sources both numbering
// from 1 must not overwrite each other in that lookup.
func TestWorkflowVariantsDoNotCollideAcrossSources(t *testing.T) {
	dir := t.TempDir()
	write := func(name, skill string, tools int) string {
		path := filepath.Join(dir, name)
		body := "{\"type\":\"session\",\"model\":\"gpt-5-codex\"}\n{\"type\":\"skill\",\"skill\":\"" + skill + "\"}\n"
		for range tools {
			body += "{\"type\":\"tool\",\"tool\":\"shell\",\"status\":\"completed\",\"payload\":{\"command\":\"echo\"}}\n"
		}
		body += "{\"type\":\"completion\",\"status\":\"success\"}\n"
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	short, long := write("short.jsonl", "shared-skill", 1), write("long.jsonl", "shared-skill", 6)

	c, err := catalog.Open(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	a := app.New(c)
	for _, in := range []string{short, long} {
		if _, err := a.Scan(context.Background(), app.ScanRequest{Input: in, Harness: "codex"}, nil); err != nil {
			t.Fatal(err)
		}
	}
	result, err := a.Analyze(context.Background(), app.AnalyzeRequest{Skill: "shared-skill", Scope: "current"})
	if err != nil {
		t.Fatal(err)
	}

	// One episode per source, and the two differ in length, so the projection
	// must report two distinct shapes rather than one shape counted twice.
	if len(result.Episodes) != 2 {
		t.Fatalf("expected one episode per source, got %d", len(result.Episodes))
	}
	if got := result.Workflow.Metrics.Variants; got != 2 {
		t.Fatalf("expected 2 distinct variants across two differently shaped sources, got %d", got)
	}
	for i, episode := range result.Episodes {
		if len(episode.EventSequences) == 0 {
			t.Fatalf("episode %d has no attributed events", i)
		}
	}
}
