package evidence

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolverVerifiesFingerprint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace")
	if err := os.WriteFile(path, []byte("safe evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	fp, _ := Fingerprint(path)
	got, err := (Resolver{}).Resolve(path, Reference{Token: "ev-random", SourceFingerprint: fp, Offset: 5, Length: 8})
	if err != nil || string(got.Content) != "evidence" {
		t.Fatalf("resolve: %q %v", got.Content, err)
	}
	if err := os.WriteFile(path, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = (Resolver{}).Resolve(path, Reference{Token: "ev-random", SourceFingerprint: fp})
	if !errors.Is(err, ErrStale) {
		t.Fatalf("expected stale, got %v", err)
	}
}
