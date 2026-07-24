package workflow

import (
	"sort"
	"strings"

	"skilltrace/internal/detection"
	"skilltrace/internal/trace"
)

type Step struct {
	Stage    Stage   `json:"stage"`
	Action   string  `json:"action"`
	Evidence []int64 `json:"evidence,omitempty"`
}

type Variant struct {
	Key      string `json:"key"`
	Steps    []Step `json:"steps"`
	Count    int    `json:"count"`
	Episodes []int  `json:"episodes"`
}

func ExactVariants(episodes []detection.Episode, events []trace.Event) []Variant {
	bySequence := make(map[int64]trace.Event, len(events))
	for _, event := range events {
		bySequence[event.Sequence] = event
	}
	variants := map[string]*Variant{}
	for episodeIndex, episode := range episodes {
		steps := make([]Step, 0, len(episode.EventSequences))
		parts := make([]string, 0, len(episode.EventSequences))
		for _, sequence := range episode.EventSequences {
			event, ok := bySequence[sequence]
			if !ok {
				continue
			}
			step := Step{Stage: Classify(event), Action: event.Kind, Evidence: []int64{sequence}}
			steps = append(steps, step)
			parts = append(parts, string(step.Stage)+":"+step.Action)
		}
		key := strings.Join(parts, ">")
		variant := variants[key]
		if variant == nil {
			variant = &Variant{Key: key, Steps: steps}
			variants[key] = variant
		}
		variant.Count++
		variant.Episodes = append(variant.Episodes, episodeIndex)
	}
	result := make([]Variant, 0, len(variants))
	for _, variant := range variants {
		result = append(result, *variant)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return result[i].Key < result[j].Key
	})
	return result
}
