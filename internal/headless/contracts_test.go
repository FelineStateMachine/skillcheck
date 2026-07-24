package headless

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestContractEnvelopeAndProgress(t *testing.T) {
	var out bytes.Buffer
	if err := Render(&out, "json", Success("discover", map[string]int{"count": 1})); err != nil {
		t.Fatal(err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope["version"] != float64(SchemaVersion) || envelope["command"] != "discover" || envelope["ok"] != true {
		t.Fatalf("unstable envelope: %#v", envelope)
	}
	out.Reset()
	if err := RenderProgress(&out, "source scan", "parse", 1, 2); err != nil {
		t.Fatal(err)
	}
	var progress Progress
	if err := json.Unmarshal(out.Bytes(), &progress); err != nil || progress.Type != "progress" || progress.Version != SchemaVersion {
		t.Fatalf("invalid progress contract: %#v, %v", progress, err)
	}
}

func TestContractExitMappings(t *testing.T) {
	cases := map[string]int{"invalid_arguments": ExitUsage, "not_found": ExitNotFound, "revision_conflict": ExitConflict, "writer_busy": ExitBusy, "cancelled": ExitCancelled, "internal_error": ExitFailure}
	for code, want := range cases {
		if got := ExitCode(code); got != want {
			t.Fatalf("ExitCode(%q) = %d, want %d", code, got, want)
		}
	}
}
