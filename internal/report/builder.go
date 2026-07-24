package report

import (
	"sort"
	"time"

	"skilltrace/internal/report/generated"
	"skilltrace/internal/workflow"
)

type BuildRequest struct {
	Comparison  workflow.Comparison
	GeneratedAt time.Time
	Freshness   string
}

func Build(request BuildRequest) generated.Dataset {
	a := newAliases()
	result := generated.Dataset{Version: "1", Title: "Skill usage comparison", GeneratedAt: request.GeneratedAt.UTC().Format(time.RFC3339), SkillAlias: "Skill 1", Freshness: request.Freshness, Caveats: []string{}}
	result.Comparison.Left = projectCohort(request.Comparison.Left, "Cohort A", a)
	result.Comparison.Right = projectCohort(request.Comparison.Right, "Cohort B", a)
	result.Comparison.ClaimsWithheld, result.Comparison.Revision = request.Comparison.ClaimsWithheld, request.Comparison.Revision
	for _, d := range request.Comparison.Differences {
		result.Comparison.Differences = append(result.Comparison.Differences, generated.Difference{ID: a.node(d.ID), Kind: d.Kind, Stage: d.Stage, Action: d.Action, Detail: d.Detail, LeftCount: d.LeftCount, RightCount: d.RightCount, Supported: d.Supported})
	}
	if request.Comparison.ClaimsWithheld {
		result.Caveats = append(result.Caveats, "Comparative claims are withheld because a cohort has limited evidence.")
	}
	return result
}

func projectCohort(c workflow.Cohort, name string, a *aliases) generated.Cohort {
	result := generated.Cohort{Alias: name, Members: len(c.Members)}
	for _, model := range c.Models {
		result.Models = append(result.Models, a.model(model.Model))
	}
	sort.Strings(result.Models)
	for _, n := range c.Graph.Nodes {
		result.Graph.Nodes = append(result.Graph.Nodes, generated.Node{ID: a.node(n.ID), Stage: string(n.Stage), Action: n.Action, Count: n.Count})
	}
	for _, e := range c.Graph.Edges {
		variants := make([]string, len(e.Variants))
		for i, v := range e.Variants {
			variants[i] = a.variant(v)
		}
		result.Graph.Edges = append(result.Graph.Edges, generated.Edge{From: a.node(e.From), To: a.node(e.To), Count: e.Count, Variants: variants})
	}
	return result
}
