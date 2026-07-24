package platform

import (
	"os"
	"path/filepath"
)

type Paths struct {
	ProjectSkills string
	GlobalSkills  string
	CodexHistory  string
}

func Resolve(project string) Paths {
	if project == "" {
		project, _ = os.Getwd()
	}
	home, _ := os.UserHomeDir()
	p := nativePaths(home)
	p.ProjectSkills = filepath.Join(project, ".codex", "skills")
	if v := os.Getenv("SKILLTRACE_SKILL_ROOT"); v != "" {
		p.GlobalSkills = v
	}
	if v := os.Getenv("SKILLTRACE_CODEX_ROOT"); v != "" {
		p.CodexHistory = v
	}
	return p
}
