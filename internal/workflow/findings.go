package workflow

import "skilltrace/internal/trace"

type Finding struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	Limited bool   `json:"limited,omitempty"`
}

func Evaluate(metrics Metrics, variants []Variant, capabilities trace.CapabilityProfile) []Finding {
	findings := []Finding{}
	if metrics.Limited() {
		findings = append(findings, Finding{ID: "limited-evidence", Kind: "data_gap", Title: "Limited evidence", Detail: "At least five uses are required for typical, unusual, better, or worse claims.", Limited: true})
	}
	if !capabilities.ToolCalls {
		findings = append(findings, Finding{ID: "tool-gap", Kind: "data_gap", Title: "Tool activity unavailable", Detail: "This source cannot support action or recovery findings."})
		return findings
	}
	if metrics.Samples >= MinimumComparativeSamples && len(variants) == 1 {
		findings = append(findings, Finding{ID: "consistent-path", Kind: "strength", Title: "Consistent observed workflow", Detail: "All attributable uses followed the same exact sequence."})
	}
	if capabilities.Usage && metrics.Samples >= MinimumComparativeSamples && metrics.Failed > 0 {
		findings = append(findings, Finding{ID: "failed-outcomes", Kind: "risk", Title: "Failed outcomes observed", Detail: "The cohort includes one or more failed outcomes."})
	}
	return findings
}
