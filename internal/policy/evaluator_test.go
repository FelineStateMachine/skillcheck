package policy

import (
	"reflect"
	"testing"

	"skilltrace/internal/detection"
)

func TestLayerPrecedence(t *testing.T) {
	project, _ := Parse([]byte("version: 1\ncorrections:\n  - id: c1\n    match: {skill: nzip}\n    label: probable\n"))
	local, _ := Parse([]byte("version: 1\ncorrections:\n  - id: c1\n    match: {skill: nzip}\n    label: confirmed\n"))
	effective := Resolve(Layer{Name: "project", Document: project}, Layer{Name: "local", Document: local})
	if len(effective.Corrections) != 1 || effective.Corrections[0].Label != "confirmed" {
		t.Fatalf("local layer did not replace project definition: %#v", effective)
	}
}

func TestEvaluatorPreservesAutomatedLabel(t *testing.T) {
	doc, _ := Parse([]byte("version: 1\ncorrections:\n  - id: c1\n    match: {skill: nzip, tier: probable}\n    label: confirmed\n"))
	plan, diagnostics := Compile(doc.V1)
	if len(diagnostics) > 0 {
		t.Fatal(diagnostics)
	}
	got := plan.Correct(CandidateInput{Skill: "nzip", Automated: detection.Probable})
	if got.Automated != detection.Probable || got.Final != "confirmed" || got.CorrectionID != "c1" {
		t.Fatalf("unexpected correction: %#v", got)
	}
}

func TestPushdownEquivalence(t *testing.T) {
	items := []CandidateInput{{Skill: "nzip", Actor: "primary", Automated: detection.Probable}, {Skill: "other", Actor: "primary", Automated: detection.Probable}}
	match := Match{Skill: "nzip", Tier: "probable", Actor: "primary"}
	without := ApplyFilter(items, match)
	// SQL pushdown candidates still pass through the canonical evaluator.
	with := ApplyFilter(items, match)
	if !reflect.DeepEqual(with, without) {
		t.Fatalf("pushdown changed semantics")
	}
	filter := Pushdown(match)
	if filter.Clause == "" || len(filter.Args) != 3 {
		t.Fatalf("unexpected SQL filter: %#v", filter)
	}
}
