package claude_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"skilltrace/internal/adapters/claude"
	"skilltrace/internal/trace"
)

func parse(t *testing.T, path string) (events []trace.Event, exclusions []trace.Exclusion, session struct {
	Project, Branch, Key, StartedAt, EndedAt string
}) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	result, err := claude.New(trace.NewHashSanitizer("test")).Parse(context.Background(), f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	session.Project = result.Session.Project
	session.Branch = result.Session.Branch
	session.Key = result.Session.Key
	session.StartedAt = result.Session.StartedAt
	session.EndedAt = result.Session.EndedAt
	return result.Events, result.Exclusions, session
}

func kinds(events []trace.Event) []string {
	out := make([]string, len(events))
	for i, e := range events {
		out[i] = e.Kind
	}
	return out
}

func TestParsesRealTranscriptShape(t *testing.T) {
	events, exclusions, session := parse(t, "../../../testdata/traces/claude/session.jsonl")

	if len(events) == 0 {
		t.Fatal("real transcript produced no events; the adapter rejected valid data")
	}

	// Session metadata: project is the bare folder, never the path.
	if session.Project != "lofi" {
		t.Fatalf("project = %q, want %q", session.Project, "lofi")
	}
	if session.Key != "SID-1" || session.Branch != "main" {
		t.Fatalf("session key/branch = %q/%q", session.Key, session.Branch)
	}
	if session.StartedAt == "" || session.EndedAt == "" || session.StartedAt == session.EndedAt {
		t.Fatalf("session should span time: %q..%q", session.StartedAt, session.EndedAt)
	}

	// The invalid trailing line is excluded, not fatal.
	var invalid bool
	for _, e := range exclusions {
		if e.Reason == "invalid_json" {
			invalid = true
		}
	}
	if !invalid {
		t.Fatalf("the malformed line should be excluded, got %+v", exclusions)
	}
}

func TestToolEventsCarryRealNames(t *testing.T) {
	events, _, _ := parse(t, "../../../testdata/traces/claude/session.jsonl")
	names := map[string]bool{}
	for _, e := range events {
		if e.Kind != "tool" {
			continue
		}
		var p trace.ToolPayload
		if err := trace.DecodePayload(e, &p); err != nil {
			t.Fatal(err)
		}
		names[p.Tool] = true
	}
	for _, want := range []string{"Read", "Edit", "Bash"} {
		if !names[want] {
			t.Fatalf("tool %q not represented; got %v", want, names)
		}
	}
	if len(names) < 3 {
		t.Fatalf("expected distinct tool names, got %v", names)
	}
}

func TestToolOutcomesArePairedByID(t *testing.T) {
	events, _, _ := parse(t, "../../../testdata/traces/claude/session.jsonl")
	statuses := map[string]int{}
	for _, e := range events {
		if e.Kind != "tool" {
			continue
		}
		var p trace.ToolPayload
		_ = trace.DecodePayload(e, &p)
		statuses[p.Status]++
	}
	if statuses["error"] == 0 {
		t.Fatalf("the failed Bash call should surface as an error outcome; statuses=%v", statuses)
	}
	if statuses["ok"] == 0 {
		t.Fatalf("successful calls should be marked ok; statuses=%v", statuses)
	}
	// Exactly one call — the trailing subagent Bash — has no result because
	// the transcript ends before it returns. That is a real shape, and it must
	// stay 'started' rather than being invented as ok.
	if statuses["started"] != 1 {
		t.Fatalf("only the unresolved trailing call should remain 'started': %v", statuses)
	}
}

// The harness's attributionSkill is what closes the "referenced but never
// invoked" gap: a skill active across turns must be detectable even when the
// Skill tool never fires.
func TestSkillDetectedFromAttribution(t *testing.T) {
	events, _, _ := parse(t, "../../../testdata/traces/claude/session.jsonl")
	tokens := []string{}
	for _, e := range events {
		if e.Kind != "skill" {
			continue
		}
		var p trace.SkillPayload
		_ = trace.DecodePayload(e, &p)
		tokens = append(tokens, p.SkillToken)
	}
	san := trace.NewHashSanitizer("test")
	frontend := san.Token("skill", "frontend-design")
	playwright := san.Token("skill", "playwright")

	// frontend-design owns four consecutive turns but must yield ONE skill
	// event, and the switch to playwright must yield a second.
	if len(tokens) != 2 {
		t.Fatalf("expected one skill event per contiguous span, got %d: %v", len(tokens), tokens)
	}
	if tokens[0] != frontend || tokens[1] != playwright {
		t.Fatalf("skill spans in wrong order: %v", tokens)
	}
}

func TestSubagentToolsAreLabelled(t *testing.T) {
	events, _, _ := parse(t, "../../../testdata/traces/claude/session.jsonl")
	var sawSubagent bool
	for _, e := range events {
		if e.Kind != "tool" {
			continue
		}
		var p trace.ToolPayload
		_ = trace.DecodePayload(e, &p)
		if p.Actor == "subagent" {
			sawSubagent = true
		}
	}
	if !sawSubagent {
		t.Fatal("the isSidechain tool call should be attributed to a subagent actor")
	}
}

func TestEmptyInputIsNotAnError(t *testing.T) {
	events, _, _ := parse(t, os.DevNull)
	if len(events) != 0 {
		t.Fatalf("empty input should yield no events, got %d", len(events))
	}
}

func TestKindsAreWithinContract(t *testing.T) {
	events, _, _ := parse(t, "../../../testdata/traces/claude/session.jsonl")
	for _, k := range kinds(events) {
		switch k {
		case "session", "skill", "tool", "file", "completion", "usage":
		default:
			t.Fatalf("event kind %q is outside the trace contract", k)
		}
	}
}

func TestIncompleteTrailingRecordIsFatal(t *testing.T) {
	// A truncated final line (no newline) means the file was captured
	// mid-write; committing it would persist half a session.
	input := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","id":"c1","name":"Bash","input":{}}]}}` + "\n" +
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use"`
	_, err := claude.New(trace.NewHashSanitizer("test")).Parse(context.Background(), strings.NewReader(input))
	if err == nil {
		t.Fatal("a truncated trailing record should be a hard error, not a silent commit")
	}
}
