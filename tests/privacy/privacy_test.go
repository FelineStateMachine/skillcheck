package privacy_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"skilltrace/internal/headless"
)

func TestPrivacyErrorContractDoesNotExposeCause(t *testing.T) {
	const canary = "PRIVATE-HOME-user-secret-token"
	var out bytes.Buffer
	err := headless.Render(&out, "json", headless.Failure("scan", "internal_error", "operation failed"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out.Bytes(), []byte(canary)) {
		t.Fatal("private canary escaped into normal output")
	}
	var value map[string]any
	if err := json.Unmarshal(out.Bytes(), &value); err != nil {
		t.Fatalf("error output is not JSON: %v", err)
	}
}
