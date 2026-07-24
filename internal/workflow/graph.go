package workflow

import "sort"

type Node struct {
	ID     string `json:"id"`
	Stage  Stage  `json:"stage"`
	Action string `json:"action"`
	Class  Class  `json:"class"`
	Count  int    `json:"count"`
	Errors int    `json:"errors,omitempty"`
}
type Edge struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	Count    int      `json:"count"`
	Variants []string `json:"variants"`
}
type Graph struct {
	Nodes    []Node    `json:"nodes"`
	Edges    []Edge    `json:"edges"`
	Variants []Variant `json:"variants"`
}

func Project(variants []Variant) Graph {
	nodes := map[string]*Node{}
	edges := map[string]*Edge{}
	for _, variant := range variants {
		previous := ""
		for index, step := range variant.Steps {
			id := string(step.Stage) + ":" + step.Action
			n := nodes[id]
			if n == nil {
				n = &Node{ID: id, Stage: step.Stage, Action: step.Action, Class: step.Class}
				nodes[id] = n
			}
			// Weight by how many episodes took this variant and how many times
			// the step repeated within it, so a node's count reflects real
			// tool volume rather than distinct appearances.
			n.Count += variant.Count * max(step.Repeat, 1)
			n.Errors += variant.Count * step.Errors
			if index > 0 {
				key := previous + ">" + id
				e := edges[key]
				if e == nil {
					e = &Edge{From: previous, To: id}
					edges[key] = e
				}
				e.Count += variant.Count
				e.Variants = append(e.Variants, variant.Key)
			}
			previous = id
		}
	}
	result := Graph{Variants: variants}
	for _, n := range nodes {
		result.Nodes = append(result.Nodes, *n)
	}
	for _, e := range edges {
		sort.Strings(e.Variants)
		result.Edges = append(result.Edges, *e)
	}
	sort.Slice(result.Nodes, func(i, j int) bool { return result.Nodes[i].ID < result.Nodes[j].ID })
	sort.Slice(result.Edges, func(i, j int) bool { a, b := result.Edges[i], result.Edges[j]; return a.From+a.To < b.From+b.To })
	return result
}

func (g Graph) HasObservedPath(nodeIDs []string) bool {
	for _, variant := range g.Variants {
		if len(variant.Steps) != len(nodeIDs) {
			continue
		}
		ok := true
		for i, s := range variant.Steps {
			if string(s.Stage)+":"+s.Action != nodeIDs[i] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}
