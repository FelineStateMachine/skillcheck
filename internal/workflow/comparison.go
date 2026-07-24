package workflow

import "sort"

type Difference struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Stage      string `json:"stage"`
	Action     string `json:"action"`
	Detail     string `json:"detail"`
	LeftCount  int    `json:"left_count"`
	RightCount int    `json:"right_count"`
	Supported  bool   `json:"supported"`
}

type Comparison struct {
	Left           Cohort        `json:"left"`
	Right          Cohort        `json:"right"`
	Compatibility  Compatibility `json:"compatibility"`
	Differences    []Difference  `json:"differences"`
	ClaimsWithheld bool          `json:"claims_withheld"`
	Revision       string        `json:"revision"`
}

func Compare(left, right Cohort) Comparison {
	result := Comparison{Left: left, Right: right, Compatibility: CompatibleCapabilities(left, right), Revision: "comparison-v1"}
	result.ClaimsWithheld = len(left.Members) < 5 || len(right.Members) < 5
	type counts struct {
		left, right int
		node        Node
	}
	all := map[string]counts{}
	for _, n := range left.Graph.Nodes {
		all[n.ID] = counts{left: n.Count, node: n}
	}
	for _, n := range right.Graph.Nodes {
		c := all[n.ID]
		c.right, c.node = n.Count, n
		all[n.ID] = c
	}
	for id, c := range all {
		if c.left == c.right {
			continue
		}
		kind := "frequency"
		if c.left == 0 {
			kind = "right_only"
		}
		if c.right == 0 {
			kind = "left_only"
		}
		detail := "observed count differs"
		supported := !result.ClaimsWithheld
		if !supported {
			detail = "difference visible; comparative claim withheld due to limited evidence"
		}
		result.Differences = append(result.Differences, Difference{ID: id, Kind: kind, Stage: string(c.node.Stage), Action: c.node.Action, Detail: detail, LeftCount: c.left, RightCount: c.right, Supported: supported})
	}
	sort.Slice(result.Differences, func(i, j int) bool { return result.Differences[i].ID < result.Differences[j].ID })
	return result
}
