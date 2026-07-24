package system_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscover(t *testing.T) {
	output := runCLI(t, "discover", "--skill-root", "testdata/skills", "--format", "json")
	if !strings.Contains(output, "\"name\":\"nzip\"") {
		t.Fatal(output)
	}
}
func TestCodexAnalysis(t *testing.T) {
	db := filepath.Join(t.TempDir(), "catalog.db")
	runCLI(t, "source", "scan", "--input", "testdata/traces/codex/minimal.jsonl", "--harness", "codex", "--catalog", db, "--format", "json")
	output := runCLI(t, "analyze", "--skill", "SECRET_SKILL_CANARY", "--scope", "current", "--catalog", db, "--format", "json")
	if !strings.Contains(output, "\"state\":\"uses\"") || strings.Contains(output, "RAW_CANARY") {
		t.Fatal(output)
	}
}
func TestNoUses(t *testing.T) {
	db := filepath.Join(t.TempDir(), "catalog.db")
	output := runCLI(t, "analyze", "--skill", "nzip", "--catalog", db, "--format", "json")
	if !strings.Contains(output, "\"state\":\"no_uses\"") {
		t.Fatal(output)
	}
}

func runCLI(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("go", append([]string{"run", "./cmd/skilltrace"}, args...)...)
	cmd.Dir = filepath.Clean("../..")
	var out, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("run: %v stderr=%s", err, stderr.String())
	}
	return out.String()
}
