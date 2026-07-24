package detection_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"skilltrace/internal/detection"
	"skilltrace/internal/trace"
)

// Attribution runs over every candidate on every analyze. Episode counts grow
// with trace volume, so the cost must stay linear in the event count.
func benchEvents(episodes int) ([]trace.Event, []detection.Candidate) {
	const eventsPerEpisode = 3
	events := make([]trace.Event, 0, episodes*eventsPerEpisode)
	candidates := make([]detection.Candidate, 0, episodes)
	sequence := int64(1)
	for i := range episodes {
		skill, _ := json.Marshal(trace.SkillPayload{SkillToken: "skill-token"})
		events = append(events, trace.Event{Sequence: sequence, Kind: "skill", Payload: skill})
		candidates = append(candidates, detection.Candidate{
			Skill: "skill-token", Tier: detection.Confirmed, Actor: "primary",
			Start: sequence, End: sequence, Evidence: []int64{sequence},
		})
		sequence++
		tool, _ := json.Marshal(map[string]string{"command": fmt.Sprintf("echo %d", i)})
		events = append(events, trace.Event{Sequence: sequence, Kind: "tool", Payload: tool})
		sequence++
		completion, _ := json.Marshal(trace.CompletionPayload{Status: "success"})
		events = append(events, trace.Event{Sequence: sequence, Kind: "completion", Payload: completion})
		sequence++
	}
	return events, candidates
}

func BenchmarkAttribute(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		events, candidates := benchEvents(n)
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = detection.Attribute(events, candidates)
			}
		})
	}
}
