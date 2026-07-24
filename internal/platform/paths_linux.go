//go:build linux

package platform

import "path/filepath"

func nativePaths(home string) Paths {
	return Paths{GlobalSkills: filepath.Join(home, ".codex", "skills"), CodexHistory: filepath.Join(home, ".codex", "sessions")}
}
