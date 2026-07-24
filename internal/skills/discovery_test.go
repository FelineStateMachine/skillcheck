package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverDeduplicatesLinkedExposure(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "real")
	if err := os.Mkdir(real, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(real, "SKILL.md"), []byte("---\nname: demo\ndescription: test\n---\nbody\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "linked")
	if err := os.Mkdir(link, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, filepath.Join(link, "demo")); err != nil {
		t.Fatal(err)
	}
	r, err := Discover([]Exposure{{Harness: "codex", Scope: "project", Root: root}, {Harness: "claude", Scope: "global", Root: link}})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Skills) != 1 || len(r.Skills[0].Exposures) != 2 {
		t.Fatalf("unexpected discovery: %#v", r)
	}
}

func TestDiscoverReportsMalformedSkill(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "bad")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("no frontmatter"), 0o600); err != nil {
		t.Fatal(err)
	}
	r, err := Discover([]Exposure{{Harness: "codex", Scope: "project", Root: root}})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Skills) != 0 || len(r.Issues) != 1 || r.Issues[0].Reason != "invalid_frontmatter" {
		t.Fatalf("unexpected result: %#v", r)
	}
}
