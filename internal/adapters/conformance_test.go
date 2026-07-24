package adapters_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"skilltrace/internal/adapters"
	"skilltrace/internal/adapters/claude"
	"skilltrace/internal/adapters/codex"
	"skilltrace/internal/adapters/huggingface"
	"skilltrace/internal/trace"
)

func TestConformanceCanonicalContract(t *testing.T) {
	sanitizer := trace.NewHashSanitizer("test")
	// Each adapter is fed one record in its own real format that carries a
	// single unit of tool activity, and must produce at least one event.
	cases := []struct {
		name  string
		input string
		parse adapters.Adapter
	}{
		{"codex", `{"type":"session","model":"gpt-5"}` + "\n", codex.New(sanitizer)},
		{"claude", `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","id":"c1","name":"Bash","input":{}}]}}` + "\n", claude.New(sanitizer)},
		{"huggingface", `{"harness":"codex","raw_retained":true,"raw_trace":[{"type":"session","model":"gpt-5"}]}` + "\n", huggingface.New(sanitizer)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.parse.Parse(context.Background(), strings.NewReader(tc.input))
			if err != nil || len(result.Events) == 0 {
				t.Fatalf("Parse() events=%d err=%v", len(result.Events), err)
			}
		})
	}
}

// Skill identities are tokenized, not labelled, so a raw skill name must never
// survive into an event.
func TestConformanceClaudeTokenizesSkillIdentity(t *testing.T) {
	input := `{"type":"assistant","attributionSkill":"secret-skill","message":{"role":"assistant","content":[{"type":"tool_use","id":"c1","name":"Bash","input":{}}]}}` + "\n"
	result, err := claude.New(trace.NewHashSanitizer("test")).Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(result.Events)
	if strings.Contains(string(encoded), "secret-skill") {
		t.Fatalf("raw skill name escaped sanitization: %s", encoded)
	}
}

// A version-gating check used to reject the semver string real Claude
// transcripts carry ("2.1.205"), discarding valid data. The adapter must now
// accept any version and parse on the record's shape instead.
func TestConformanceClaudeAcceptsAnyVersion(t *testing.T) {
	for _, version := range []string{`"2.1.205"`, `1`, `null`} {
		input := `{"type":"assistant","version":` + version + `,"message":{"role":"assistant","content":[{"type":"tool_use","id":"c1","name":"Read","input":{}}]}}` + "\n"
		result, err := claude.New(trace.NewHashSanitizer("test")).Parse(context.Background(), strings.NewReader(input))
		if err != nil {
			t.Fatalf("version %s: %v", version, err)
		}
		if len(result.Events) != 1 {
			t.Fatalf("version %s: expected the record to parse, got %d events", version, len(result.Events))
		}
	}
}
