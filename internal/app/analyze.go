package app

import (
	"context"
	"skilltrace/internal/detection"
	"skilltrace/internal/trace"
)

type AnalyzeRequest struct{ Skill, Scope string }
type AnalyzeResult struct {
	Skill    string                `json:"skill"`
	Scope    string                `json:"scope"`
	State    string                `json:"state"`
	Episodes []detection.Episode   `json:"episodes"`
	Possible []detection.Candidate `json:"possible,omitempty"`
	Workflow WorkflowResult        `json:"workflow"`
}

func (a *Application) Analyze(ctx context.Context, req AnalyzeRequest) (AnalyzeResult, error) {
	events, err := a.Catalog.CurrentEvents(ctx)
	if err != nil {
		return AnalyzeResult{}, err
	}
	token := trace.NewHashSanitizer("skilltrace-v1").Token("skill", req.Skill)
	candidates := detection.Candidates(events, token)
	accepted := make([]detection.Candidate, 0, len(candidates))
	var possible []detection.Candidate
	for _, c := range candidates {
		if c.Tier == detection.Possible {
			possible = append(possible, c)
		} else {
			accepted = append(accepted, c)
		}
	}
	episodes := detection.Attribute(events, accepted)
	if err := a.Catalog.SaveEpisodes(ctx, req.Skill, episodes); err != nil {
		return AnalyzeResult{}, err
	}
	state := "uses"
	if len(episodes) == 0 {
		state = "no_uses"
	}
	workflowResult, err := a.BuildWorkflow(ctx, episodes)
	if err != nil {
		return AnalyzeResult{}, err
	}
	return AnalyzeResult{Skill: req.Skill, Scope: req.Scope, State: state, Episodes: episodes, Possible: possible, Workflow: workflowResult}, nil
}
