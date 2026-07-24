package system_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceScan(t *testing.T) {
	root := filepath.Clean("../..")
	db := filepath.Join(t.TempDir(), "catalog.db")
	cmd := exec.Command("go", "run", "./cmd/skilltrace", "source", "scan", "--input", "testdata/traces/codex/minimal.jsonl", "--harness", "codex", "--format", "json", "--catalog", db)
	cmd.Dir = root
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("run: %v stderr=%s", err, stderr.String())
	}
	var envelope struct {
		Version int  `json:"version"`
		OK      bool `json:"ok"`
		Data    struct {
			Source struct {
				Status     string `json:"status"`
				EventCount int    `json:"event_count"`
				Revision   int    `json:"revision"`
			} `json:"source"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Version != 1 || !envelope.OK || envelope.Data.Source.Status != "supported" || envelope.Data.Source.EventCount != 6 || envelope.Data.Source.Revision != 1 {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
	assertNoCanaries(t, stdout.Bytes())
	database, err := os.ReadFile(db)
	if err != nil {
		t.Fatal(err)
	}
	assertNoCanaries(t, database)
}

func assertNoCanaries(t *testing.T, b []byte) {
	t.Helper()
	for _, canary := range []string{"RAW_CANARY", "SECRET_SKILL_CANARY", "/Users/private"} {
		if strings.Contains(string(b), canary) {
			t.Fatalf("privacy canary leaked: %s", canary)
		}
	}
}
