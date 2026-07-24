package policy

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBytePreservation(t *testing.T) {
	original := []byte("# retained\r\nversion: 1\r\ncorrections:\r\n  - id: c1\r\n    enabled: true # retained too\r\n    match: {skill: nzip}\r\n    label: confirmed\r\n")
	updated, err := SetEnabled(original, "correction", "c1", false)
	if err != nil {
		t.Fatal(err)
	}
	want := bytes.Replace(original, []byte("enabled: true"), []byte("enabled: false"), 1)
	if !bytes.Equal(updated, want) {
		t.Fatalf("unrelated bytes changed\nwant %q\n got %q", want, updated)
	}
}

func TestConflict(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy.yaml")
	initial := []byte("version: 1\n")
	if err := os.WriteFile(path, initial, 0600); err != nil {
		t.Fatal(err)
	}
	doc, diagnostics, err := (Store{}).Read(path)
	if err != nil || len(diagnostics) > 0 {
		t.Fatalf("read: %v %v", err, diagnostics)
	}
	if err := os.WriteFile(path, []byte("version: 1\n# external\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := (Store{}).Replace(path, doc.Fingerprint, initial); err == nil {
		t.Fatal("expected fingerprint conflict")
	}
}

func TestAliasesAreReadOnly(t *testing.T) {
	original := []byte("version: 1\ncorrections:\n  - &base\n    id: c1\n    match: {}\n    label: confirmed\n")
	if _, err := SetEnabled(original, "correction", "c1", false); err == nil {
		t.Fatal("expected alias document to be read-only")
	}
}
