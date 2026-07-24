package app

import (
	"context"
	"sort"

	"skilltrace/internal/trace"
	"skilltrace/internal/workflow"
)

type WorkflowResult struct {
	Graph        workflow.Graph          `json:"graph"`
	Metrics      workflow.Metrics        `json:"metrics"`
	Findings     []workflow.Finding      `json:"findings"`
	Capabilities trace.CapabilityProfile `json:"capabilities"`
}

// BuildWorkflow projects episodes onto the events they were attributed from.
//
// Projection is done per session and then merged. Doing it over a combined
// event slice would key the sequence lookup across sessions that each number
// from one, so one session's events would silently replace another's. The
// caller supplies the events because loading them is the dominant cost of an
// analyze, and the analyze path has already materialised them.
func (a *Application) BuildWorkflow(ctx context.Context, analyses []sessionAnalysis) (WorkflowResult, error) {
	var (
		metrics      workflow.Metrics
		capabilities trace.CapabilityProfile
		merged       = map[string]*workflow.Variant{}
		offset       int
	)
	for _, analysis := range analyses {
		for _, variant := range workflow.ExactVariants(analysis.episodes, analysis.events) {
			existing := merged[variant.Key]
			if existing == nil {
				clone := variant
				clone.Episodes = nil
				existing = &clone
				merged[variant.Key] = existing
			}
			existing.Count += variant.Count
			// Episode indices are local to a session; shift them so they
			// address the combined episode list the caller sees.
			for _, index := range variant.Episodes {
				existing.Episodes = append(existing.Episodes, offset+index)
			}
		}
		offset += len(analysis.episodes)
		for _, episode := range analysis.episodes {
			switch episode.Outcome {
			case "succeeded":
				metrics.Completed++
			case "failed":
				metrics.Failed++
			default:
				metrics.Unknown++
			}
		}
		capabilities = mergeCapabilities(capabilities, capabilitiesFromEvents(analysis.events))
	}

	variants := sortedVariants(merged)
	metrics.Samples = offset
	metrics.Variants = len(variants)
	result := WorkflowResult{Graph: workflow.Project(variants), Metrics: metrics, Findings: workflow.Evaluate(metrics, variants, capabilities), Capabilities: capabilities}
	if err := a.Catalog.SaveWorkflow(ctx, result.Graph, result.Metrics, result.Findings); err != nil {
		return WorkflowResult{}, err
	}
	return result, nil
}

// sortedVariants orders merged variants the same way ExactVariants does for a
// single session: most frequent first, ties broken by key so output is stable.
func sortedVariants(merged map[string]*workflow.Variant) []workflow.Variant {
	out := make([]workflow.Variant, 0, len(merged))
	for _, variant := range merged {
		out = append(out, *variant)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Key < out[j].Key
	})
	return out
}

// mergeCapabilities unions what each session was able to observe. A capability
// present in any session is available to the analysis as a whole.
func mergeCapabilities(a, b trace.CapabilityProfile) trace.CapabilityProfile {
	return trace.CapabilityProfile{
		Actors:        a.Actors || b.Actors,
		SkillEvidence: a.SkillEvidence || b.SkillEvidence,
		ToolCalls:     a.ToolCalls || b.ToolCalls,
		FileEvents:    a.FileEvents || b.FileEvents,
		Usage:         a.Usage || b.Usage,
	}
}

func capabilitiesFromEvents(events []trace.Event) trace.CapabilityProfile {
	var p trace.CapabilityProfile
	for _, event := range events {
		switch event.Kind {
		case "session":
			p.Actors = true
		case "skill":
			p.SkillEvidence = true
		case "tool":
			p.ToolCalls = true
		case "file":
			p.FileEvents = true
		case "usage":
			p.Usage = true
		}
	}
	return p
}
