package system

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"skilltrace/internal/detection"
	"skilltrace/internal/policy"
)

func TestPolicyPreviewApply(t *testing.T) {
	root := filepath.Clean("../..")
	catalogPath := filepath.Join(t.TempDir(), "catalog.db")
	preview := exec.Command("go", "run", "./cmd/skilltrace", "policy", "preview", "--file", "testdata/policy/valid.yaml", "--catalog", catalogPath, "--format", "json")
	preview.Dir = root
	out, err := preview.Output()
	if err != nil {
		t.Fatalf("preview failed: %v", err)
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &envelope); err != nil || !envelope.OK || envelope.Data.Token == "" {
		t.Fatalf("unexpected preview: %v %s", err, out)
	}
	apply := exec.Command("go", "run", "./cmd/skilltrace", "policy", "apply", "--file", "testdata/policy/valid.yaml", "--catalog", catalogPath, "--token", envelope.Data.Token, "--format", "json")
	apply.Dir = root
	if out, err = apply.Output(); err != nil {
		t.Fatalf("apply failed: %v %s", err, out)
	}
}

func TestCorrectionReclassifies(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("../..", "testdata/policy/valid.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	doc, diagnostics := policy.Parse(data)
	if len(diagnostics) != 0 || doc == nil {
		t.Fatalf("valid policy rejected: %v", diagnostics)
	}
	plan, diagnostics := policy.Compile(doc.V1)
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	result := plan.Correct(policy.CandidateInput{Skill: "nzip", Automated: detection.Probable})
	if result.Automated != detection.Probable || result.Final != "confirmed" {
		t.Fatalf("unexpected correction result: %#v", result)
	}
}
