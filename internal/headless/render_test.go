package headless_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"skilltrace/internal/headless"
)

func TestJSONEnvelope(t *testing.T) {
	var b bytes.Buffer
	if err := headless.Render(&b, "json", headless.Success("source health", []string{})); err != nil {
		t.Fatal(err)
	}
	var got headless.Envelope
	if err := json.Unmarshal(b.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Version != 1 || !got.OK || got.Command != "source health" {
		t.Fatalf("unexpected envelope: %+v", got)
	}
}
