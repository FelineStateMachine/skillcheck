package app

import "skilltrace/internal/skills"

type DiscoverRequest struct{ SkillRoot string }
type DiscoverResult struct {
	Skills []skills.Skill `json:"skills"`
	Issues []skills.Issue `json:"issues,omitempty"`
}

func (a *Application) Discover(req DiscoverRequest) (DiscoverResult, error) {
	r, err := skills.Discover([]skills.Exposure{{Harness: "codex", Scope: "project", Root: req.SkillRoot}})
	return DiscoverResult{Skills: r.Skills, Issues: r.Issues}, err
}
