package skills

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"skilltrace/internal/text"
)

func Discover(roots []Exposure) (Result, error) {
	byID := map[string]*Skill{}
	var result Result
	for _, exposure := range roots {
		entries, err := os.ReadDir(exposure.Root)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return result, fmt.Errorf("read skill root: %w", err)
		}
		for _, entry := range entries {
			path, err := resolve(filepath.Join(exposure.Root, entry.Name()))
			if err != nil {
				result.Issues = append(result.Issues, Issue{entry.Name(), "invalid_link"})
				continue
			}
			b, err := os.ReadFile(filepath.Join(path, "SKILL.md"))
			if err != nil {
				result.Issues = append(result.Issues, Issue{entry.Name(), "missing_skill_document"})
				continue
			}
			name, desc, ok := frontmatter(string(b))
			if !ok {
				result.Issues = append(result.Issues, Issue{entry.Name(), "invalid_frontmatter"})
				continue
			}
			id := identity(name, desc, string(b))
			s := byID[id.ID]
			if s == nil {
				s = &Skill{Identity: id}
				byID[id.ID] = s
			}
			s.Exposures = append(s.Exposures, Exposure{Harness: exposure.Harness, Scope: exposure.Scope, Root: filepath.Base(exposure.Root)})
		}
	}
	for _, s := range byID {
		sort.Slice(s.Exposures, func(i, j int) bool {
			if s.Exposures[i].Scope != s.Exposures[j].Scope {
				return s.Exposures[i].Scope == "project"
			}
			return s.Exposures[i].Harness < s.Exposures[j].Harness
		})
		result.Skills = append(result.Skills, *s)
	}
	sort.Slice(result.Skills, func(i, j int) bool {
		if result.Skills[i].Identity.Name != result.Skills[j].Identity.Name {
			return result.Skills[i].Identity.Name < result.Skills[j].Identity.Name
		}
		return result.Skills[i].Identity.ID < result.Skills[j].Identity.ID
	})
	return result, nil
}

func resolve(path string) (string, error) {
	for i := 0; i < maxLinkDepth; i++ {
		info, err := os.Lstat(path)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink == 0 {
			return path, nil
		}
		next, err := os.Readlink(path)
		if err != nil {
			return "", err
		}
		if !filepath.IsAbs(next) {
			next = filepath.Join(filepath.Dir(path), next)
		}
		path = filepath.Clean(next)
	}
	return "", fmt.Errorf("link depth exceeded")
}
func frontmatter(doc string) (string, string, bool) {
	s := bufio.NewScanner(strings.NewReader(doc))
	if !s.Scan() || strings.TrimSpace(s.Text()) != "---" {
		return "", "", false
	}
	vals := map[string]string{}
	closed := false
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "---" {
			closed = true
			break
		}
		p := strings.SplitN(line, ":", 2)
		if len(p) == 2 {
			vals[strings.TrimSpace(p[0])] = strings.Trim(strings.TrimSpace(p[1]), "\"'")
		}
	}
	// Names and descriptions come from third-party skill documents and are
	// rendered straight into a terminal, so control sequences are stripped at
	// ingest rather than at each display site.
	name := text.Sanitize(vals["name"])
	return name, text.Sanitize(vals["description"]), closed && name != ""
}
