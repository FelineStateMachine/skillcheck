package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

var supportedHarnesses = map[string]bool{"codex": true, "claude": true, "huggingface": true}

type Root struct {
	Harness string `json:"harness"`
	Path    string `json:"path"`
}
type Settings struct {
	Version int    `json:"version"`
	Roots   []Root `json:"roots"`
}

func ValidateRoot(root Root) (Root, error) {
	if !supportedHarnesses[root.Harness] {
		return Root{}, fmt.Errorf("unsupported harness")
	}
	path, err := filepath.Abs(root.Path)
	if err != nil {
		return Root{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return Root{}, err
	}
	if !info.IsDir() {
		return Root{}, fmt.Errorf("custom root is not a directory")
	}
	root.Path = filepath.Clean(path)
	return root, nil
}

func (s *Settings) Add(root Root) error {
	root, err := ValidateRoot(root)
	if err != nil {
		return err
	}
	for _, existing := range s.Roots {
		if existing == root {
			return nil
		}
	}
	s.Roots = append(s.Roots, root)
	sort.Slice(s.Roots, func(i, j int) bool {
		if s.Roots[i].Harness != s.Roots[j].Harness {
			return s.Roots[i].Harness < s.Roots[j].Harness
		}
		return s.Roots[i].Path < s.Roots[j].Path
	})
	return nil
}

func (s *Settings) Forget(root Root) bool {
	for i, existing := range s.Roots {
		if existing == root {
			s.Roots = append(s.Roots[:i], s.Roots[i+1:]...)
			return true
		}
	}
	return false
}
