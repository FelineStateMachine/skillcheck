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
	cases := []struct {
		name  string
		input string
		parse adapters.Adapter
	}{
		{"codex", `{"type":"session","model":"gpt-5"}` + "\n", codex.New(sanitizer)},
		{"claude", `{"type":"session_start","version":1,"model":"claude-sonnet-4","actor":"main"}` + "\n", claude.New(sanitizer)},
		{"huggingface", `{"harness":"codex","raw_retained":true,"raw_trace":[{"type":"session","model":"gpt-5"}]}` + "\n", huggingface.New(sanitizer)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.parse.Parse(context.Background(), strings.NewReader(tc.input))
			if err != nil || len(result.Events) != 1 {
				t.Fatalf("Parse() events=%d err=%v", len(result.Events), err)
			}
			encoded, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(encoded), "main") {
				t.Fatal("raw actor identity escaped sanitization")
			}
		})
	}
}

func TestConformanceUnknownClaudeVersionIsExcluded(t *testing.T) {
	result, err := claude.New(trace.NewHashSanitizer("test")).Parse(context.Background(), strings.NewReader(`{"type":"session","version":99}`+"\n"))
	if err != nil || len(result.Events) != 0 || len(result.Exclusions) != 1 {
		t.Fatalf("unexpected result: %#v, %v", result, err)
	}
}
