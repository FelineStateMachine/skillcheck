package workflow

import (
	"fmt"
	"sort"
	"time"

	"skilltrace/internal/trace"
)

type ModelScope string

const (
	ExactModel  ModelScope = "exact"
	ModelFamily ModelScope = "family"
	AllModels   ModelScope = "all"
)

type CohortDefinition struct {
	ID         string     `json:"id"`
	Label      string     `json:"label"`
	Skill      string     `json:"skill"`
	Harness    string     `json:"harness,omitempty"`
	Model      string     `json:"model,omitempty"`
	Family     string     `json:"family,omitempty"`
	ModelScope ModelScope `json:"model_scope"`
	Repository string     `json:"repository,omitempty"`
	Actor      string     `json:"actor,omitempty"`
	From       string     `json:"from,omitempty"`
	To         string     `json:"to,omitempty"`
}

type CohortMember struct {
	ID           string                  `json:"id"`
	Skill        string                  `json:"skill"`
	Harness      string                  `json:"harness"`
	Model        string                  `json:"model"`
	Family       string                  `json:"family,omitempty"`
	Repository   string                  `json:"repository,omitempty"`
	Actor        string                  `json:"actor,omitempty"`
	OccurredAt   time.Time               `json:"occurred_at"`
	Variant      Variant                 `json:"variant"`
	Capabilities trace.CapabilityProfile `json:"capabilities"`
}

type ModelCount struct {
	Model string `json:"model"`
	Count int    `json:"count"`
}

type Cohort struct {
	Definition   CohortDefinition        `json:"definition"`
	Members      []CohortMember          `json:"members"`
	Models       []ModelCount            `json:"models"`
	Capabilities trace.CapabilityProfile `json:"capabilities"`
	Graph        Graph                   `json:"graph"`
	Warnings     []string                `json:"warnings,omitempty"`
}

func BuildCohort(def CohortDefinition, candidates []CohortMember) (Cohort, error) {
	if def.Skill == "" {
		return Cohort{}, fmt.Errorf("cohort skill is required")
	}
	if def.ModelScope == "" {
		def.ModelScope = ExactModel
	}
	if def.ModelScope == ExactModel && def.Model == "" {
		return Cohort{}, fmt.Errorf("exact model cohort requires model")
	}
	from, err := parseDate(def.From, false)
	if err != nil {
		return Cohort{}, fmt.Errorf("invalid from date: %w", err)
	}
	to, err := parseDate(def.To, true)
	if err != nil {
		return Cohort{}, fmt.Errorf("invalid to date: %w", err)
	}
	cohort := Cohort{Definition: def}
	modelCounts := map[string]int{}
	for _, member := range candidates {
		if !matches(def, member, from, to) {
			continue
		}
		cohort.Members = append(cohort.Members, member)
		modelCounts[member.Model]++
	}
	sort.Slice(cohort.Members, func(i, j int) bool { return cohort.Members[i].ID < cohort.Members[j].ID })
	variants := make([]Variant, 0, len(cohort.Members))
	for _, member := range cohort.Members {
		variants = append(variants, member.Variant)
	}
	cohort.Graph = Project(aggregateVariants(variants))
	for model, count := range modelCounts {
		cohort.Models = append(cohort.Models, ModelCount{Model: model, Count: count})
	}
	sort.Slice(cohort.Models, func(i, j int) bool { return cohort.Models[i].Model < cohort.Models[j].Model })
	cohort.Capabilities = commonCapabilities(cohort.Members)
	if len(cohort.Models) > 1 {
		cohort.Warnings = append(cohort.Warnings, "mixed model cohort")
	}
	if len(cohort.Members) < 5 {
		cohort.Warnings = append(cohort.Warnings, "limited evidence: fewer than 5 uses")
	}
	if len(cohort.Members) == 0 {
		cohort.Warnings = append(cohort.Warnings, "empty cohort")
	}
	return cohort, nil
}

func matches(d CohortDefinition, m CohortMember, from, to time.Time) bool {
	if m.Skill != d.Skill || d.Harness != "" && m.Harness != d.Harness || d.Repository != "" && m.Repository != d.Repository || d.Actor != "" && m.Actor != d.Actor {
		return false
	}
	switch d.ModelScope {
	case ExactModel:
		if m.Model != d.Model {
			return false
		}
	case ModelFamily:
		if m.Family != d.Family {
			return false
		}
	case AllModels:
	default:
		return false
	}
	return (from.IsZero() || !m.OccurredAt.Before(from)) && (to.IsZero() || !m.OccurredAt.After(to))
}

func parseDate(value string, end bool) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err == nil && end {
		t = t.Add(24*time.Hour - time.Nanosecond)
	}
	return t, err
}

func aggregateVariants(input []Variant) []Variant {
	byKey := map[string]*Variant{}
	for _, v := range input {
		item := byKey[v.Key]
		if item == nil {
			copy := v
			copy.Count = 0
			item = &copy
			byKey[v.Key] = item
		}
		count := v.Count
		if count < 1 {
			count = 1
		}
		item.Count += count
	}
	result := make([]Variant, 0, len(byKey))
	for _, v := range byKey {
		result = append(result, *v)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return result[i].Key < result[j].Key
	})
	return result
}

func commonCapabilities(members []CohortMember) trace.CapabilityProfile {
	if len(members) == 0 {
		return trace.CapabilityProfile{}
	}
	p := members[0].Capabilities
	for _, m := range members[1:] {
		p.Actors = p.Actors && m.Capabilities.Actors
		p.SkillEvidence = p.SkillEvidence && m.Capabilities.SkillEvidence
		p.ToolCalls = p.ToolCalls && m.Capabilities.ToolCalls
		p.FileEvents = p.FileEvents && m.Capabilities.FileEvents
		p.Usage = p.Usage && m.Capabilities.Usage
	}
	return p
}
