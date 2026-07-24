package detection

import (
	"sort"

	"skilltrace/internal/trace"
)

type Episode struct {
	Skill          string  `json:"skill"`
	Actor          string  `json:"actor"`
	Tier           Tier    `json:"tier"`
	Start          int64   `json:"start"`
	End            int64   `json:"end"`
	EventSequences []int64 `json:"event_sequences"`
	Outcome        string  `json:"outcome"`
}

// Attribute assigns events to the candidate whose window contains them.
//
// Candidate windows are half-open against the next candidate's start, so when
// candidates are ordered by start they are disjoint and ascending. That lets a
// single pass over the events place every event with one binary search, which
// keeps attribution linear in the event count. Candidates are produced in event
// order by Candidates, so the ordered path is the one taken in practice;
// unordered input falls back to the exhaustive scan so results never depend on
// how the caller assembled its candidates.
func Attribute(events []trace.Event, candidates []Candidate) []Episode {
	if !ascendingByStart(candidates) {
		return attributeExhaustive(events, candidates)
	}

	ends := make([]int64, len(candidates))
	for i := range candidates {
		switch {
		case i+1 < len(candidates):
			ends[i] = candidates[i+1].Start - 1
		case len(events) > 0:
			ends[i] = events[len(events)-1].Sequence
		}
	}

	sequences := make([][]int64, len(candidates))
	for i := range sequences {
		sequences[i] = []int64{}
	}
	outcomes := make([]string, len(candidates))
	// Windows are disjoint, so a sequence belongs to at most one of them and a
	// single set is enough to deduplicate repeated sequences across all windows.
	seen := make(map[int64]bool, len(events))

	for _, event := range events {
		window := windowFor(candidates, ends, event.Sequence)
		if window < 0 {
			continue
		}
		if !seen[event.Sequence] {
			seen[event.Sequence] = true
			sequences[window] = append(sequences[window], event.Sequence)
		}
		if outcomes[window] == "" && event.Kind == "completion" {
			outcomes[window] = completionOutcome(event)
		}
	}

	result := make([]Episode, 0, len(candidates))
	for i, c := range candidates {
		outcome := outcomes[i]
		if outcome == "" {
			outcome = "unknown"
		}
		result = append(result, Episode{
			Skill: c.Skill, Actor: c.Actor, Tier: c.Tier, Start: c.Start, End: ends[i],
			EventSequences: sequences[i], Outcome: outcome,
		})
	}
	return result
}

func ascendingByStart(candidates []Candidate) bool {
	for i := 1; i < len(candidates); i++ {
		if candidates[i-1].Start > candidates[i].Start {
			return false
		}
	}
	return true
}

// windowFor returns the index of the candidate window holding sequence, or -1
// when the sequence falls before the first candidate or past the last window.
func windowFor(candidates []Candidate, ends []int64, sequence int64) int {
	i := sort.Search(len(candidates), func(i int) bool { return candidates[i].Start > sequence }) - 1
	if i < 0 || sequence > ends[i] {
		return -1
	}
	return i
}

func attributeExhaustive(events []trace.Event, candidates []Candidate) []Episode {
	result := make([]Episode, 0, len(candidates))
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
		result = append(result, Episode{Skill: c.Skill, Actor: c.Actor, Tier: c.Tier, Start: c.Start, End: end, EventSequences: seq, Outcome: Outcome(events, c.Start, end)})
	}
	return result
}
