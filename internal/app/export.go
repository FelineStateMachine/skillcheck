package app

import (
	"context"
	"fmt"
	"path/filepath"

	"skilltrace/internal/report"
)

type ExportRequest struct{ Skill, LeftPath, RightPath, Output string }
type ExportResult struct {
	Output   string `json:"output"`
	Format   string `json:"format"`
	Revision string `json:"revision"`
}

func (a *Application) ExportHTML(ctx context.Context, request ExportRequest) (ExportResult, error) {
	if request.Output == "" {
		return ExportResult{}, fmt.Errorf("output is required")
	}
	comparison, err := a.Compare(ctx, CompareRequest{Skill: request.Skill, LeftPath: request.LeftPath, RightPath: request.RightPath})
	if err != nil {
		return ExportResult{}, err
	}
	dataset := report.Build(report.BuildRequest{Comparison: comparison.Comparison, GeneratedAt: comparison.GeneratedAt, Freshness: comparison.Freshness})
	output, err := filepath.Abs(request.Output)
	if err != nil {
		return ExportResult{}, err
	}
	if err := report.Export(output, dataset); err != nil {
		return ExportResult{}, fmt.Errorf("export html: %w", err)
	}
	return ExportResult{Output: output, Format: "html", Revision: comparison.CatalogRevision}, nil
}
