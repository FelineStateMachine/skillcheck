package workflow

import (
	"testing"
	"time"

	"skilltrace/internal/trace"
)

func member(id, model, family, key string, count int) CohortMember {
	return CohortMember{ID: id, Skill: "nzip", Harness: "codex", Model: model, Family: family, OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Capabilities: trace.CapabilityProfile{ToolCalls: true, Usage: true}, Variant: Variant{Key: key, Count: count, Steps: []Step{{Stage: StageAct, Action: key}}}}
}

func TestCohortExactAndMixedModels(t *testing.T) {
	candidates := []CohortMember{member("b", "gpt-5", "gpt", "edit", 1), member("a", "gpt-5.1", "gpt", "test", 1), member("c", "gpt-5", "gpt", "edit", 1)}
	exact, err := BuildCohort(CohortDefinition{Skill: "nzip", ModelScope: ExactModel, Model: "gpt-5"}, candidates)
	if err != nil {
		t.Fatal(err)
	}
	if len(exact.Members) != 2 || exact.Members[0].ID != "b" {
		t.Fatalf("unexpected exact cohort: %#v", exact.Members)
	}
	mixed, err := BuildCohort(CohortDefinition{Skill: "nzip", ModelScope: ModelFamily, Family: "gpt"}, candidates)
	if err != nil {
		t.Fatal(err)
	}
	if len(mixed.Models) != 2 || mixed.Warnings[0] != "mixed model cohort" {
		t.Fatalf("missing composition warning: %#v", mixed)
	}
}

func TestComparisonWithholdsClaimsAndPreservesGraphs(t *testing.T) {
	left, _ := BuildCohort(CohortDefinition{Skill: "nzip", ModelScope: ExactModel, Model: "gpt-5"}, []CohortMember{member("a", "gpt-5", "gpt", "edit", 1)})
	rightMember := member("b", "claude", "claude", "test", 1)
	rightMember.Capabilities.Usage = false
	right, _ := BuildCohort(CohortDefinition{Skill: "nzip", ModelScope: ExactModel, Model: "claude"}, []CohortMember{rightMember})
	c := Compare(left, right)
	if !c.ClaimsWithheld || c.Compatibility.Usage || len(c.Left.Graph.Nodes) != 1 || len(c.Right.Graph.Nodes) != 1 {
		t.Fatalf("unexpected comparison: %#v", c)
	}
	for _, d := range c.Differences {
		if d.Supported {
			t.Fatal("limited evidence difference was presented as supported")
		}
	}
}
