package system_test

import (
	"path/filepath"
	"strings"
	"testing"

	"skilltrace/internal/app"
	"skilltrace/internal/catalog"
	"skilltrace/internal/tui"
)

func TestTUIDiscovery(t *testing.T) {
	c, err := catalog.Open(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	a := app.New(c)
	root, err := tui.Load(a, tui.Config{SkillRoot: "../../testdata/skills", Unicode: false})
	if err != nil {
		t.Fatal(err)
	}
	view := root.View().Content
	if !strings.Contains(view, "nzip") || !strings.Contains(view, "Sources") {
		t.Fatalf("representative discovery view is incomplete:\n%s", view)
	}
}
