package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"skilltrace/internal/workflow"
)

type CohortFile struct {
	Definition workflow.CohortDefinition `json:"definition"`
	Members    []workflow.CohortMember   `json:"members"`
}

type CompareRequest struct{ Skill, LeftPath, RightPath string }
type CompareResult struct {
	Skill           string              `json:"skill"`
	Comparison      workflow.Comparison `json:"comparison"`
	GeneratedAt     time.Time           `json:"generated_at"`
	CatalogRevision string              `json:"catalog_revision"`
	Freshness       string              `json:"freshness"`
}

func (a *Application) Compare(ctx context.Context, request CompareRequest) (CompareResult, error) {
	leftFile, err := readCohortFile(request.LeftPath)
	if err != nil {
		return CompareResult{}, fmt.Errorf("read left cohort: %w", err)
	}
	rightFile, err := readCohortFile(request.RightPath)
	if err != nil {
		return CompareResult{}, fmt.Errorf("read right cohort: %w", err)
	}
	if request.Skill == "" {
		return CompareResult{}, fmt.Errorf("skill is required")
	}
	if leftFile.Definition.Skill != request.Skill || rightFile.Definition.Skill != request.Skill {
		return CompareResult{}, fmt.Errorf("cohorts must select skill %q", request.Skill)
	}
	left, err := workflow.BuildCohort(leftFile.Definition, leftFile.Members)
	if err != nil {
		return CompareResult{}, err
	}
	right, err := workflow.BuildCohort(rightFile.Definition, rightFile.Members)
	if err != nil {
		return CompareResult{}, err
	}
	comparison := workflow.Compare(left, right)
	if err := a.Catalog.SaveComparison(ctx, comparison); err != nil {
		return CompareResult{}, err
	}
	return CompareResult{Skill: request.Skill, Comparison: comparison, GeneratedAt: time.Unix(0, 0).UTC(), CatalogRevision: comparison.Revision, Freshness: "current"}, nil
}

func readCohortFile(path string) (CohortFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return CohortFile{}, err
	}
	var result CohortFile
	if err := json.Unmarshal(b, &result); err != nil {
		return CohortFile{}, fmt.Errorf("decode cohort definition (files use JSON-compatible YAML): %w", err)
	}
	return result, nil
}
