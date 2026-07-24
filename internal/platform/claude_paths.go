package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

func ClaudeRoots() []string {
	if root := os.Getenv("SKILLTRACE_CLAUDE_ROOT"); root != "" {
		return []string{root}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	if runtime.GOOS == "windows" {
		return []string{filepath.Join(home, ".claude", "projects")}
	}
	return []string{filepath.Join(home, ".claude", "projects")}
}
