package app

import (
	"context"

	"skilltrace/internal/detection"
	"skilltrace/internal/trace"
	"skilltrace/internal/workflow"
)

type WorkflowResult struct {
	Graph        workflow.Graph          `json:"graph"`
	Metrics      workflow.Metrics        `json:"metrics"`
	Findings     []workflow.Finding      `json:"findings"`
	Capabilities trace.CapabilityProfile `json:"capabilities"`
}

func (a *Application) BuildWorkflow(ctx context.Context, episodes []detection.Episode) (WorkflowResult, error) {
	events, err := a.Catalog.CurrentEvents(ctx)
	if err != nil {
		return WorkflowResult{}, err
	}
	variants := workflow.ExactVariants(episodes, events)
	metrics := workflow.Metrics{Samples: len(episodes), Variants: len(variants)}
	for _, episode := range episodes {
		switch episode.Outcome {
		case "succeeded":
			metrics.Completed++
		case "failed":
			metrics.Failed++
		default:
			metrics.Unknown++
		}
	}
	capabilities := capabilitiesFromEvents(events)
	result := WorkflowResult{Graph: workflow.Project(variants), Metrics: metrics, Findings: workflow.Evaluate(metrics, variants, capabilities), Capabilities: capabilities}
	if err := a.Catalog.SaveWorkflow(ctx, result.Graph, result.Metrics, result.Findings); err != nil {
		return WorkflowResult{}, err
	}
	return result, nil
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
