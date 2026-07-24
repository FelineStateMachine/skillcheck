package detection_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"skilltrace/internal/detection"
	"skilltrace/internal/trace"
)

// referenceAttribute is the original O(candidates x events) implementation,
// kept here so the optimised version can be proven to produce identical
// output rather than merely plausible output.
func referenceAttribute(events []trace.Event, candidates []detection.Candidate) []detection.Episode {
	result := make([]detection.Episode, 0, len(candidates))
	for i, c := range candidates {
		end := int64(0)
		if i+1 < len(candidates) {
			end = candidates[i+1].Start - 1
		} else if len(events) > 0 {
			end = events[len(events)-1].Sequence
		}
		seq := []int64{}
		seen := map[int64]bool{}
		for _, e := range events {
			if e.Sequence >= c.Start && e.Sequence <= end && !seen[e.Sequence] {
				seq = append(seq, e.Sequence)
				seen[e.Sequence] = true
			}
		}
		result = append(result, detection.Episode{
			Skill: c.Skill, Actor: c.Actor, Tier: c.Tier, Start: c.Start, End: end,
			EventSequences: seq, Outcome: detection.Outcome(events, c.Start, end),
		})
	}
	return result
}

func payload(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestAttributeMatchesReferenceImplementation(t *testing.T) {
	cases := map[string]struct {
		events     []trace.Event
		candidates []detection.Candidate
	}{
		"empty": {},
		"no candidates": {
			events: []trace.Event{{Sequence: 1, Kind: "tool"}},
		},
		"candidates without events": {
			candidates: []detection.Candidate{{Skill: "s", Start: 1, End: 1}},
		},
		"single episode succeeded": {
			events: []trace.Event{
				{Sequence: 1, Kind: "skill"},
				{Sequence: 2, Kind: "tool"},
				{Sequence: 3, Kind: "completion", Payload: payload(t, trace.CompletionPayload{Status: "success"})},
			},
			candidates: []detection.Candidate{{Skill: "s", Actor: "primary", Tier: detection.Confirmed, Start: 1, End: 1}},
		},
		"adjacent episodes split at boundary": {
			events: []trace.Event{
				{Sequence: 1, Kind: "skill"},
				{Sequence: 2, Kind: "completion", Payload: payload(t, trace.CompletionPayload{Status: "failed"})},
				{Sequence: 3, Kind: "skill"},
				{Sequence: 4, Kind: "completion", Payload: payload(t, trace.CompletionPayload{Status: "completed"})},
			},
			candidates: []detection.Candidate{
				{Skill: "s", Actor: "primary", Tier: detection.Confirmed, Start: 1, End: 1},
				{Skill: "s", Actor: "primary", Tier: detection.Confirmed, Start: 3, End: 3},
			},
		},
		"sparse and duplicated sequences": {
			events: []trace.Event{
				{Sequence: 1, Kind: "skill"},
				{Sequence: 1, Kind: "tool"},
				{Sequence: 5, Kind: "tool"},
				{Sequence: 9, Kind: "completion", Payload: payload(t, trace.CompletionPayload{Status: "unknown-status"})},
				{Sequence: 20, Kind: "tool"},
			},
			candidates: []detection.Candidate{
				{Skill: "s", Actor: "primary", Tier: detection.Confirmed, Start: 1, End: 1},
				{Skill: "s", Actor: "sub", Tier: detection.Probable, Start: 9, End: 9},
			},
		},
		"candidate starts after every event": {
			events: []trace.Event{
				{Sequence: 1, Kind: "skill"},
				{Sequence: 2, Kind: "tool"},
			},
			candidates: []detection.Candidate{{Skill: "s", Actor: "primary", Tier: detection.Confirmed, Start: 7, End: 7}},
		},
		"undecodable completion payload": {
			events: []trace.Event{
				{Sequence: 1, Kind: "skill"},
				{Sequence: 2, Kind: "completion", Payload: json.RawMessage("not json")},
			},
			candidates: []detection.Candidate{{Skill: "s", Actor: "primary", Tier: detection.Confirmed, Start: 1, End: 1}},
		},
		"first completion in window wins": {
			events: []trace.Event{
				{Sequence: 1, Kind: "skill"},
				{Sequence: 2, Kind: "completion", Payload: payload(t, trace.CompletionPayload{Status: "failure"})},
				{Sequence: 3, Kind: "completion", Payload: payload(t, trace.CompletionPayload{Status: "success"})},
			},
			candidates: []detection.Candidate{{Skill: "s", Actor: "primary", Tier: detection.Confirmed, Start: 1, End: 1}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			want := referenceAttribute(tc.events, tc.candidates)
			got := detection.Attribute(tc.events, tc.candidates)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("attribution diverged from the reference implementation\n got: %#v\nwant: %#v", got, want)
			}
		})
	}
}

func TestAttributeMatchesReferenceOnGeneratedTraces(t *testing.T) {
	for _, n := range []int{1, 2, 17, 250} {
		events, candidates := benchEvents(n)
		want := referenceAttribute(events, candidates)
		got := detection.Attribute(events, candidates)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("attribution diverged at %d episodes", n)
		}
	}
}
