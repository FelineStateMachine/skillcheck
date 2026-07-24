package skills

import "path/filepath"

type Exposure struct {
	Harness string `json:"harness"`
	Scope   string `json:"scope"`
	Root    string `json:"root"`
}
type Skill struct {
	Identity  Identity   `json:"identity"`
	Exposures []Exposure `json:"exposures"`
}
type Issue struct {
	Entry  string `json:"entry"`
	Reason string `json:"reason"`
}
type Result struct {
	Skills []Skill `json:"skills"`
	Issues []Issue `json:"issues,omitempty"`
}

// DefaultExposures lists where skills are installed for the harnesses this
// tool understands. Searching only the project-local Codex directory made the
// common case — a user with skills installed globally, running in an unrelated
// repository — indistinguishable from having no skills at all.
//
// Discover skips roots that do not exist and merges skills that appear under
// more than one, so listing every candidate here is safe.
func DefaultExposures(workingDir, homeDir string) []Exposure {
	var exposures []Exposure
	if workingDir != "" {
		exposures = append(exposures,
			Exposure{Harness: "codex", Scope: "project", Root: filepath.Join(workingDir, ".codex", "skills")},
			Exposure{Harness: "claude", Scope: "project", Root: filepath.Join(workingDir, ".claude", "skills")},
		)
	}
	if homeDir != "" {
		exposures = append(exposures,
			Exposure{Harness: "codex", Scope: "global", Root: filepath.Join(homeDir, ".codex", "skills")},
			Exposure{Harness: "claude", Scope: "global", Root: filepath.Join(homeDir, ".claude", "skills")},
		)
	}
	return exposures
}
