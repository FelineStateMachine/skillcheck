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
	Class    Class   `json:"class"`
	Repeat   int     `json:"repeat"`           // collapsed consecutive occurrences (>=1)
	Errors   int     `json:"errors,omitempty"` // how many of those occurrences failed
	Evidence []int64 `json:"evidence,omitempty"`
}

type Variant struct {
	Key      string `json:"key"`
	Steps    []Step `json:"steps"`
	Count    int    `json:"count"`
	Episodes []int  `json:"episodes"`
}

// ExactVariants groups episodes by the shape of their tool sequence.
//
// Two things make the shape legible on real data. Node identity is the concrete
// tool name, not the bare kind "tool", so Bash and Read are distinct nodes.
// And consecutive repeats of the same action are collapsed into one step with a
// count, because real sessions are dominated by runs — Bash after Bash after
// Bash — that would otherwise swamp the graph and make every session unique.
func ExactVariants(episodes []detection.Episode, events []trace.Event) []Variant {
	bySequence := make(map[int64]trace.Event, len(events))
	for _, event := range events {
		bySequence[event.Sequence] = event
	}
	variants := map[string]*Variant{}
	for episodeIndex, episode := range episodes {
		steps := collapse(stepsForEpisode(episode, bySequence))
		key := variantKey(steps)
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

func stepsForEpisode(episode detection.Episode, bySequence map[int64]trace.Event) []Step {
	steps := make([]Step, 0, len(episode.EventSequences))
	for _, sequence := range episode.EventSequences {
		event, ok := bySequence[sequence]
		if !ok {
			continue
		}
		step := Step{
			Stage: Classify(event), Action: Action(event), Class: classify(event),
			Repeat: 1, Evidence: []int64{sequence},
		}
		if errored(event) {
			step.Errors = 1
		}
		steps = append(steps, step)
	}
	return steps
}

// collapse folds consecutive steps with the same action into one, summing their
// repeat and error counts and keeping a bounded sample of evidence.
func collapse(steps []Step) []Step {
	const maxEvidence = 8
	out := make([]Step, 0, len(steps))
	for _, step := range steps {
		if n := len(out); n > 0 && out[n-1].Stage == step.Stage && out[n-1].Action == step.Action {
			out[n-1].Repeat += step.Repeat
			out[n-1].Errors += step.Errors
			if len(out[n-1].Evidence) < maxEvidence {
				out[n-1].Evidence = append(out[n-1].Evidence, step.Evidence...)
			}
			continue
		}
		out = append(out, step)
	}
	return out
}

func variantKey(steps []Step) string {
	parts := make([]string, len(steps))
	for i, step := range steps {
		parts[i] = string(step.Stage) + ":" + step.Action
	}
	return strings.Join(parts, ">")
}
