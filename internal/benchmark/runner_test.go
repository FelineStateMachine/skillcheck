package benchmark

import (
	"context"
	"path/filepath"
	"testing"
)

func TestRunnerUsesProductionAdapters(t *testing.T) {
	report, err := Run(context.Background(), filepath.Join("..", "..", "testdata", "benchmark"))
	if err != nil {
		t.Fatal(err)
	}
	if report.Fixtures != 1 || len(report.Slices) != 1 || report.Slices[0].Recall != 1 {
		t.Fatalf("unexpected report: %#v", report)
	}
}
