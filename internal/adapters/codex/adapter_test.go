package codex_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"skilltrace/internal/adapters/codex"
	"skilltrace/internal/trace"
)

func TestAdapterNormalizesAndSanitizes(t *testing.T) {
	f, err := os.Open("../../../testdata/traces/codex/minimal.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	result, err := codex.New(trace.NewHashSanitizer("test")).Parse(context.Background(), f)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Events) != 6 || len(result.Exclusions) != 1 {
		t.Fatalf("events=%d exclusions=%d", len(result.Events), len(result.Exclusions))
	}
	for _, e := range result.Events {
		if strings.Contains(string(e.Payload), "CANARY") || strings.Contains(string(e.Payload), "/Users/") {
			t.Fatalf("raw value leaked: %s", e.Payload)
		}
	}
}

// A malformed line is excluded, not fatal: one bad record in a long real
// session must not discard everything around it.
func TestAdapterExcludesMalformedRecordWithoutAborting(t *testing.T) {
	result, err := codex.New(trace.NewHashSanitizer("test")).Parse(context.Background(),
		strings.NewReader("{bad}\n{\"type\":\"session\",\"model\":\"gpt\"}\n"))
	if err != nil {
		t.Fatalf("a malformed line should not abort the parse: %v", err)
	}
	if len(result.Events) != 1 {
		t.Fatalf("the valid record after the bad one should survive: %d events", len(result.Events))
	}
	if len(result.Exclusions) != 1 || result.Exclusions[0].Reason != "invalid_json" {
		t.Fatalf("the bad line should be one invalid_json exclusion: %+v", result.Exclusions)
	}
}

func TestReadsRealRolloutFormat(t *testing.T) {
	f, err := os.Open("../../../testdata/traces/codex/session.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	result, err := codex.New(trace.NewHashSanitizer("test")).Parse(context.Background(), f)
	if err != nil {
		t.Fatal(err)
	}

	if result.Session.Project != "puzzletea" {
		t.Fatalf("project = %q, want bare folder %q", result.Session.Project, "puzzletea")
	}
	if result.Session.Branch == "" {
		t.Fatal("branch should be captured from session_meta.git")
	}

	var tools, ok, errored int
	for _, e := range result.Events {
		if e.Kind != "tool" {
			continue
		}
		tools++
		var p trace.ToolPayload
		if err := trace.DecodePayload(e, &p); err != nil {
			t.Fatal(err)
		}
		switch p.Status {
		case "ok":
			ok++
		case "error":
			errored++
		}
	}
	if tools == 0 {
		t.Fatal("no tool events extracted from a real rollout")
	}
	if ok == 0 || errored == 0 {
		t.Fatalf("outcomes should be paired from function_call_output: ok=%d error=%d", ok, errored)
	}
	for _, e := range result.Events {
		if strings.Contains(string(e.Payload), "CANARY") || strings.Contains(string(e.Payload), "/Users/") {
			t.Fatalf("raw command text or path leaked into an event: %s", e.Payload)
		}
	}
}
