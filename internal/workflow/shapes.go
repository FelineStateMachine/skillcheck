package workflow

import (
	"sort"
	"strings"
)

// Shape is a class-level summary of a workflow: the sequence of semantic
// classes a group of episodes followed, with how many took it. Concrete tools
// vary too much to group whole sequences, but their classes recur, so this is
// the layer at which "the usual workflow" becomes legible.
type Shape struct {
	Classes []Class `json:"classes"`
	Count   int     `json:"count"`
}

func (s Shape) String() string {
	parts := make([]string, len(s.Classes))
	for i, c := range s.Classes {
		parts[i] = string(c)
	}
	return strings.Join(parts, " → ")
}

// Shapes ranks the class-level shapes across all variants, most common first.
// Runs of one class are collapsed so "execute execute execute" reads as a
// single execute stage.
func Shapes(variants []Variant, limit int) []Shape {
	counts := map[string]*Shape{}
	for _, variant := range variants {
		classes := normalizeClasses(variant.Steps)
		key := shapeKey(classes)
		shape := counts[key]
		if shape == nil {
			shape = &Shape{Classes: classes}
			counts[key] = shape
		}
		shape.Count += variant.Count
	}
	out := make([]Shape, 0, len(counts))
	for _, shape := range counts {
		out = append(out, *shape)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return shapeKey(out[i].Classes) < shapeKey(out[j].Classes)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func normalizeClasses(steps []Step) []Class {
	out := make([]Class, 0, len(steps))
	for _, step := range steps {
		if n := len(out); n > 0 && out[n-1] == step.Class {
			continue
		}
		out = append(out, step.Class)
	}
	return out
}

func shapeKey(classes []Class) string {
	parts := make([]string, len(classes))
	for i, c := range classes {
		parts[i] = string(c)
	}
	return strings.Join(parts, ">")
}
