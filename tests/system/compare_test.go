package system

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCompare(t *testing.T) {
	root := filepath.Clean("../..")
	cmd := exec.Command("go", "run", "./cmd/skilltrace", "compare", "--skill", "nzip", "--left", "testdata/cohorts/codex.yaml", "--right", "testdata/cohorts/claude.yaml", "--catalog", filepath.Join(t.TempDir(), "catalog.db"), "--format", "json")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("compare failed: %v", err)
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Skill string `json:"skill"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &envelope); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if !envelope.OK || envelope.Data.Skill != "nzip" {
		t.Fatalf("unexpected result: %s", out)
	}
}
