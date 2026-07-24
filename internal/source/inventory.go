package source

import (
	"path/filepath"
	"sort"
)

type Candidate struct {
	Harness, Path             string
	CurrentRepository, Custom bool
}

func Inventory(currentRoot string, native, custom map[string][]string) []Candidate {
	var out []Candidate
	add := func(h string, paths []string, customRoot bool) {
		for _, p := range paths {
			clean, _ := filepath.Abs(p)
			root, _ := filepath.Abs(currentRoot)
			rel, err := filepath.Rel(root, clean)
			current := err == nil && rel != ".." && rel != "." && len(rel) > 0 && rel[:1] != string(filepath.Separator)
			out = append(out, Candidate{Harness: h, Path: clean, CurrentRepository: current, Custom: customRoot})
		}
	}
	for h, paths := range native {
		add(h, paths, false)
	}
	for h, paths := range custom {
		add(h, paths, true)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CurrentRepository != out[j].CurrentRepository {
			return out[i].CurrentRepository
		}
		if out[i].Harness != out[j].Harness {
			return out[i].Harness < out[j].Harness
		}
		return out[i].Path < out[j].Path
	})
	return out
}
