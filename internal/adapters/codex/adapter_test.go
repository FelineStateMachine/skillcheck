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

func TestAdapterRejectsMalformedRecord(t *testing.T) {
	_, err := codex.New(trace.NewHashSanitizer("test")).Parse(context.Background(), strings.NewReader("{bad}\n"))
	if err == nil {
		t.Fatal("expected malformed input error")
	}
}
