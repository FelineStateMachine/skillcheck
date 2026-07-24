package policy

import "skilltrace/internal/detection"

type SQLFilter struct {
	Clause string
	Args   []any
}

// Pushdown lowers only equality predicates over indexed, catalog-owned fields.
func Pushdown(match Match) SQLFilter {
	filter := SQLFilter{Clause: "1=1"}
	add := func(column, value string) {
		if value != "" {
			filter.Clause += " AND " + column + " = ?"
			filter.Args = append(filter.Args, value)
		}
	}
	add("skill_name", match.Skill)
	add("automated_label", match.Tier)
	add("actor", match.Actor)
	return filter
}

func ApplyFilter(items []CandidateInput, match Match) []CandidateInput {
	var result []CandidateInput
	for _, item := range items {
		if matches(match, item, string(item.Automated)) {
			result = append(result, item)
		}
	}
	return result
}

func Tier(value string) detection.Tier { return detection.Tier(value) }
