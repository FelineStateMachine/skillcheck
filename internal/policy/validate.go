package policy

import "fmt"

func Validate(d DocumentV1) []Diagnostic {
	var out []Diagnostic
	if d.Version != 1 {
		out = append(out, Diagnostic{Code: "unsupported_version", Message: "policy version must be 1"})
	}
	seen := map[string]string{}
	check := func(kind, id string) {
		if id == "" {
			out = append(out, Diagnostic{Code: "missing_id", Message: kind + " id is required"})
			return
		}
		key := kind + ":" + id
		if _, ok := seen[key]; ok {
			out = append(out, Diagnostic{Code: "duplicate_id", Message: fmt.Sprintf("duplicate %s id %q", kind, id)})
		}
		seen[key] = kind
	}
	for _, c := range d.Corrections {
		check("correction", c.ID)
		if c.Label != "confirmed" && c.Label != "probable" && c.Label != "possible" && c.Label != "excluded" {
			out = append(out, Diagnostic{Code: "invalid_label", Message: "correction label must be confirmed, probable, possible, or excluded"})
		}
	}
	for _, f := range d.Findings {
		check("finding", f.ID)
		if f.Kind == "" || f.Title == "" || f.Detail == "" {
			out = append(out, Diagnostic{Code: "invalid_finding", Message: "finding kind, title, and detail are required"})
		}
	}
	for _, c := range d.Cohorts {
		check("cohort", c.ID)
	}
	for _, r := range d.ActionRollups {
		check("action_rollup", r.ID)
		if r.Stage == "" || len(r.Actions) == 0 {
			out = append(out, Diagnostic{Code: "invalid_rollup", Message: "action rollup stage and actions are required"})
		}
	}
	return out
}
