package source

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestAppendEquivalence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	if err := os.WriteFile(path, []byte("{\"n\":1}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := FingerprintFile(path, 8)
	if err != nil {
		t.Fatal(err)
	}
	cp := Checkpoint{Fingerprint: before, CommittedOffset: 8, ParserRevision: "v1"}
	if err := os.WriteFile(path, []byte("{\"n\":1}\n{\"n\":2}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	after, err := FingerprintFile(path, 8)
	if err != nil {
		t.Fatal(err)
	}
	if err := cp.AppendEligible(after, "v1"); err != nil {
		t.Fatal(err)
	}
	r, offset, err := OpenSuffix(path, 8)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	b, _ := io.ReadAll(r)
	if string(b) != "{\"n\":2}\n" || offset != 16 {
		t.Fatalf("suffix=%q offset=%d", b, offset)
	}
}

func TestTrailingRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	if err := os.WriteFile(path, []byte("one\ntwo"), 0o600); err != nil {
		t.Fatal(err)
	}
	r, offset, err := OpenSuffix(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	b, _ := io.ReadAll(r)
	if string(b) != "one\n" || offset != 4 {
		t.Fatalf("suffix=%q offset=%d", b, offset)
	}
}

func TestDigestMismatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	if err := os.WriteFile(path, []byte("first\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fp, _ := FingerprintFile(path, 6)
	cp := Checkpoint{Fingerprint: fp, CommittedOffset: 6, ParserRevision: "v1"}
	if err := os.WriteFile(path, []byte("other\nnext\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	changed, _ := FingerprintFile(path, 6)
	if cp.AppendEligible(changed, "v1") == nil {
		t.Fatal("expected digest mismatch")
	}
}
