package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"skilltrace/internal/workflow"
)

func TestReportContractEscapesAndAliasesIdentities(t *testing.T) {
	canary := "</script><script>fetch('https://secret.example')</script>/Users/private"
	cohort := workflow.Cohort{Members: []workflow.CohortMember{{ID: canary}}, Models: []workflow.ModelCount{{Model: canary, Count: 1}}, Graph: workflow.Graph{Nodes: []workflow.Node{{ID: canary, Stage: workflow.StageAct, Action: "run", Count: 1}}}}
	dataset := Build(BuildRequest{Comparison: workflow.Comparison{Left: cohort, Right: cohort}, GeneratedAt: time.Unix(0, 0), Freshness: "current"})
	path := filepath.Join(t.TempDir(), "report.html")
	if err := Export(path, dataset); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if strings.Contains(text, canary) || strings.Contains(text, "/Users/private") || strings.Contains(text, "https://") {
		t.Fatalf("report leaked identity or network reference")
	}
	if !strings.Contains(text, "report-data") || !strings.Contains(text, "model-1") {
		t.Fatalf("report is missing embedded dataset")
	}
}

func TestReportContractDeterministicDataset(t *testing.T) {
	a := Build(BuildRequest{GeneratedAt: time.Unix(0, 0), Freshness: "current"})
	b := Build(BuildRequest{GeneratedAt: time.Unix(0, 0), Freshness: "current"})
	if a.GeneratedAt != b.GeneratedAt || a.Version != "1" {
		t.Fatal("dataset is not deterministic or versioned")
	}
}
