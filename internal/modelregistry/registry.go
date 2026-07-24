package modelregistry

import "strings"

type Identity struct{ Exact, Provider, Family string }

// Resolve derives display groupings while preserving the exact source identity.
func Resolve(exact string) Identity {
	lower := strings.ToLower(exact)
	i := Identity{Exact: exact, Family: exact}
	switch {
	case strings.Contains(lower, "claude"):
		i.Provider = "anthropic"
		i.Family = "claude"
	case strings.Contains(lower, "gpt") || strings.Contains(lower, "o1") || strings.Contains(lower, "o3"):
		i.Provider = "openai"
		i.Family = "openai-reasoning"
	}
	return i
}
