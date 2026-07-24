package detection

import "skilltrace/internal/trace"

type Episode struct {
	Skill          string  `json:"skill"`
	Actor          string  `json:"actor"`
	Tier           Tier    `json:"tier"`
	Start          int64   `json:"start"`
	End            int64   `json:"end"`
	EventSequences []int64 `json:"event_sequences"`
	Outcome        string  `json:"outcome"`
}

func Attribute(events []trace.Event, candidates []Candidate) []Episode {
	result := make([]Episode, 0, len(candidates))
	for i, c := range candidates {
		end := int64(0)
		if i+1 < len(candidates) {
			end = candidates[i+1].Start - 1
		} else if len(events) > 0 {
			end = events[len(events)-1].Sequence
		}
		seq := []int64{}
		for _, e := range events {
			if e.Sequence >= c.Start && e.Sequence <= end {
				seq = append(seq, e.Sequence)
			}
		}
		result = append(result, Episode{Skill: c.Skill, Actor: c.Actor, Tier: c.Tier, Start: c.Start, End: end, EventSequences: seq, Outcome: Outcome(events, c.Start, end)})
	}
	return result
}
