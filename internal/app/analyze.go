package app

import (
	"context"

	"skilltrace/internal/catalog"
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

// sessionAnalysis keeps each session's events beside the episodes attributed
// from them, so every downstream sequence lookup stays inside one session.
type sessionAnalysis struct {
	session  catalog.Session
	events   []trace.Event
	episodes []detection.Episode
}

func (a *Application) Analyze(ctx context.Context, req AnalyzeRequest) (AnalyzeResult, error) {
	sessions, err := a.Catalog.CurrentSessions(ctx)
	if err != nil {
		return AnalyzeResult{}, err
	}
	token := trace.NewHashSanitizer("skilltrace-v1").Token("skill", req.Skill)

	analyses := make([]sessionAnalysis, 0, len(sessions))
	var episodes []detection.Episode
	var possible []detection.Candidate
	for _, session := range sessions {
		candidates := detection.Candidates(session.Events, token)
		accepted := make([]detection.Candidate, 0, len(candidates))
		for _, c := range candidates {
			if c.Tier == detection.Possible {
				possible = append(possible, c)
			} else {
				accepted = append(accepted, c)
			}
		}
		// Attribution windows are bounded by this session's own events, so a
		// trailing episode cannot run on into the next session's stream.
		attributed := detection.Attribute(session.Events, accepted)
		if len(attributed) == 0 {
			continue
		}
		analyses = append(analyses, sessionAnalysis{session: session.Session, events: session.Events, episodes: attributed})
		episodes = append(episodes, attributed...)
	}

	if err := a.Catalog.SaveEpisodes(ctx, req.Skill, episodes); err != nil {
		return AnalyzeResult{}, err
	}
	state := "uses"
	if len(episodes) == 0 {
		state = "no_uses"
	}
	workflowResult, err := a.BuildWorkflow(ctx, analyses)
	if err != nil {
		return AnalyzeResult{}, err
	}
	return AnalyzeResult{Skill: req.Skill, Scope: req.Scope, State: state, Episodes: episodes, Possible: possible, Workflow: workflowResult}, nil
}
