package app

import (
	"os"

	"skilltrace/internal/skills"
)

// DiscoverRequest selects where to look for installed skills. An explicit
// SkillRoot pins the search to exactly that directory; leaving it empty
// searches the default project and global locations for every known harness.
type DiscoverRequest struct{ SkillRoot string }

type DiscoverResult struct {
	Skills []skills.Skill `json:"skills"`
	Issues []skills.Issue `json:"issues,omitempty"`
}

func (r DiscoverRequest) exposures() []skills.Exposure {
	if r.SkillRoot != "" {
		// A pinned root says nothing about which harness installed it, so it
		// is reported as pinned rather than guessed as Codex.
		return []skills.Exposure{{Scope: "pinned", Root: r.SkillRoot}}
	}
	workingDir, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()
	return skills.DefaultExposures(workingDir, homeDir)
}

// Roots lists the directories this request will search, so an empty result can
// report where it looked.
func (r DiscoverRequest) Roots() []string {
	exposures := r.exposures()
	roots := make([]string, 0, len(exposures))
	for _, exposure := range exposures {
		roots = append(roots, exposure.Root)
	}
	return roots
}

func (a *Application) Discover(req DiscoverRequest) (DiscoverResult, error) {
	r, err := skills.Discover(req.exposures())
	return DiscoverResult{Skills: r.Skills, Issues: r.Issues}, err
}
