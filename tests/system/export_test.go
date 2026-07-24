package system

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHTMLExport(t *testing.T) {
	root := filepath.Clean("../..")
	dir := t.TempDir()
	output := filepath.Join(dir, "comparison.html")
	cmd := exec.Command("go", "run", "./cmd/skilltrace", "export", "html", "--skill", "nzip", "--left", "testdata/cohorts/codex.yaml", "--right", "testdata/cohorts/claude.yaml", "--catalog", filepath.Join(dir, "catalog.db"), "--output", output, "--format", "json")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Format string `json:"format"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &envelope); err != nil || !envelope.OK || envelope.Data.Format != "html" {
		t.Fatalf("unexpected result: %s", out)
	}
	b, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	html := string(b)
	if !strings.Contains(html, "report-data") || !strings.Contains(html, "<!doctype html>") {
		t.Fatal("export is not a self-contained HTML report")
	}
	if strings.Contains(html, "https://") || strings.Contains(html, "http://") || strings.Contains(html, "nzip") {
		t.Fatal("export contains a network reference or original skill identity")
	}
}
