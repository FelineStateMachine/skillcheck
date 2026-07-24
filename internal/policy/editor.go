package policy

import (
	"bytes"
	"fmt"
	"regexp"
)

type ConflictError struct{ Message string }

func (e ConflictError) Error() string { return e.Message }

// SetEnabled performs a source-range splice on a plain mapping entry. Aliased
// or merged entries are deliberately read-only because their edit target is
// ambiguous. The resulting complete document is always revalidated.
func SetEnabled(original []byte, kind, id string, value bool) ([]byte, error) {
	if bytes.Contains(original, []byte("<<:")) || bytes.Contains(original, []byte("&")) || bytes.Contains(original, []byte("*")) {
		return nil, ConflictError{Message: "policy uses YAML aliases or merges and is read-only"}
	}
	section := sectionName(kind)
	if section == "" {
		return nil, fmt.Errorf("unsupported policy kind")
	}
	lines := splitLines(original)
	start, end := -1, len(lines)
	sectionRE := regexp.MustCompile(`^` + regexp.QuoteMeta(section) + `:\s*(?:#.*)?(?:\r?\n)?$`)
	for i, line := range lines {
		if start < 0 && sectionRE.Match(line) {
			start = i + 1
			continue
		}
		if start >= 0 && len(line) > 0 && line[0] != ' ' && line[0] != '\t' && line[0] != '#' {
			end = i
			break
		}
	}
	if start < 0 {
		return nil, fmt.Errorf("policy section not found")
	}
	idRE := regexp.MustCompile(`^\s*-\s+id:\s*["']?` + regexp.QuoteMeta(id) + `["']?\s*(?:#.*)?(?:\r?\n)?$`)
	itemStart, itemEnd := -1, end
	for i := start; i < end; i++ {
		if idRE.Match(lines[i]) {
			itemStart = i
			continue
		}
		if itemStart >= 0 && regexp.MustCompile(`^\s*-\s+id:`).Match(lines[i]) {
			itemEnd = i
			break
		}
	}
	if itemStart < 0 {
		return nil, fmt.Errorf("policy definition not found")
	}
	enabledRE := regexp.MustCompile(`^(\s*)enabled:\s*(?:true|false)(\s*(?:#.*)?\r?\n?)$`)
	replacement := []byte(fmt.Sprintf("enabled: %t", value))
	for i := itemStart + 1; i < itemEnd; i++ {
		if match := enabledRE.FindSubmatch(lines[i]); match != nil {
			lines[i] = append(append(append([]byte{}, match[1]...), replacement...), match[2]...)
			return validatedJoin(lines)
		}
	}
	indent := []byte("    ")
	if m := regexp.MustCompile(`^(\s*)-`).FindSubmatch(lines[itemStart]); m != nil {
		indent = append(append([]byte{}, m[1]...), "  "...)
	}
	ending := []byte("\n")
	if bytes.HasSuffix(lines[itemStart], []byte("\r\n")) {
		ending = []byte("\r\n")
	}
	line := append(append(append(indent, replacement...), ending...), []byte{}...)
	lines = append(lines[:itemStart+1], append([][]byte{line}, lines[itemStart+1:]...)...)
	return validatedJoin(lines)
}

func sectionName(kind string) string {
	switch kind {
	case "correction":
		return "corrections"
	case "finding":
		return "findings"
	case "cohort":
		return "cohorts"
	case "action_rollup":
		return "action_rollups"
	}
	return ""
}
func splitLines(data []byte) [][]byte {
	parts := bytes.SplitAfter(data, []byte("\n"))
	if len(parts) > 0 && len(parts[len(parts)-1]) == 0 {
		parts = parts[:len(parts)-1]
	}
	return parts
}
func validatedJoin(lines [][]byte) ([]byte, error) {
	result := bytes.Join(lines, nil)
	if _, diagnostics := Parse(result); len(diagnostics) > 0 {
		return nil, diagnostics[0]
	}
	return result, nil
}
