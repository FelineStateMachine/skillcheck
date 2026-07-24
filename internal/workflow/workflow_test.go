package workflow

import (
	"encoding/json"
	"testing"

	"skilltrace/internal/detection"
	"skilltrace/internal/trace"
)

func event(t *testing.T, kind string, sequence int64) trace.Event {
	t.Helper()
	e, err := trace.NewEvent(kind, sequence, sequence, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestPathProjectionDoesNotInventCompletePath(t *testing.T) {
	events := []trace.Event{event(t, "skill", 1), event(t, "tool", 2), event(t, "completion", 3), event(t, "skill", 4), event(t, "usage", 5), event(t, "completion", 6)}
	episodes := []detection.Episode{{EventSequences: []int64{1, 2, 3}}, {EventSequences: []int64{4, 5, 6}}}
	graph := Project(ExactVariants(episodes, events))
	if graph.HasObservedPath([]string{"invoke:skill", "act:tool", "verify:usage", "finish:completion"}) {
		t.Fatal("projected graph accepted an unobserved recombined path")
	}
	if !graph.HasObservedPath([]string{"invoke:skill", "act:tool", "finish:completion"}) {
		t.Fatal("observed path missing")
	}
}

func TestPathStableTieOrdering(t *testing.T) {
	variants := ExactVariants([]detection.Episode{{EventSequences: []int64{1}}, {EventSequences: []int64{2}}}, []trace.Event{event(t, "tool", 1), event(t, "skill", 2)})
	if len(variants) != 2 || variants[0].Key > variants[1].Key {
		t.Fatalf("unstable variants: %#v", variants)
	}
}

func TestCapabilityGatesFindings(t *testing.T) {
	got := Evaluate(Metrics{Samples: 8, Failed: 2}, nil, trace.CapabilityProfile{})
	for _, finding := range got {
		if finding.ID == "failed-outcomes" {
			t.Fatal("outcome finding emitted without required capabilities")
		}
	}
}

func TestLimitedEvidenceWithholdsComparativeClaims(t *testing.T) {
	got := Evaluate(Metrics{Samples: 4}, []Variant{{}}, trace.CapabilityProfile{ToolCalls: true, Usage: true})
	b, _ := json.Marshal(got)
	if len(got) != 1 || got[0].ID != "limited-evidence" {
		t.Fatalf("unexpected limited findings: %s", b)
	}
}
