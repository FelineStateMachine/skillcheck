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

func toolEvent(t *testing.T, name, status string, sequence int64) trace.Event {
	t.Helper()
	e, err := trace.NewEvent("tool", sequence, sequence, trace.ToolPayload{Tool: name, Status: status})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// The old graph labelled every tool "tool" and had at most six nodes. Distinct
// tools must now be distinct nodes.
func TestNodesUseConcreteToolNames(t *testing.T) {
	events := []trace.Event{toolEvent(t, "Bash", "ok", 1), toolEvent(t, "Read", "ok", 2)}
	episodes := []detection.Episode{{EventSequences: []int64{1, 2}}}
	graph := Project(ExactVariants(episodes, events))
	names := map[string]bool{}
	for _, n := range graph.Nodes {
		names[n.Action] = true
	}
	if !names["Bash"] || !names["Read"] {
		t.Fatalf("expected Bash and Read nodes, got %v", names)
	}
}

// Real sessions are dominated by runs; consecutive repeats must collapse so a
// long Bash run is one step, not a unique shape.
func TestConsecutiveRepeatsCollapse(t *testing.T) {
	events := []trace.Event{toolEvent(t, "Bash", "ok", 1), toolEvent(t, "Bash", "error", 2), toolEvent(t, "Bash", "ok", 3)}
	episodes := []detection.Episode{{EventSequences: []int64{1, 2, 3}}}
	variants := ExactVariants(episodes, events)
	if len(variants) != 1 || len(variants[0].Steps) != 1 {
		t.Fatalf("three Bash calls should collapse to one step: %#v", variants)
	}
	step := variants[0].Steps[0]
	if step.Repeat != 3 || step.Errors != 1 {
		t.Fatalf("collapsed step should record 3 repeats and 1 error: %#v", step)
	}
}

// Two sessions that both go Bash then Read share one class-level shape even
// when their exact tool counts differ.
func TestShapesGroupByClass(t *testing.T) {
	events := []trace.Event{
		toolEvent(t, "Bash", "ok", 1), toolEvent(t, "Read", "ok", 2),
		toolEvent(t, "Bash", "ok", 3), toolEvent(t, "Bash", "ok", 4), toolEvent(t, "Grep", "ok", 5),
	}
	episodes := []detection.Episode{{EventSequences: []int64{1, 2}}, {EventSequences: []int64{3, 4, 5}}}
	shapes := Shapes(ExactVariants(episodes, events), 6)
	if len(shapes) != 1 {
		t.Fatalf("execute->inspect should be one shape for both episodes: %#v", shapes)
	}
	if shapes[0].Count != 2 {
		t.Fatalf("shape should count both episodes, got %d", shapes[0].Count)
	}
	if got := shapes[0].String(); got != "execute → inspect" {
		t.Fatalf("shape string = %q", got)
	}
}
