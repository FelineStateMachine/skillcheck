package policy

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type Plan struct {
	Revision      string
	Corrections   []Correction
	Findings      []FindingRule
	Cohorts       []Cohort
	ActionRollups []ActionRollup
}

func Compile(document DocumentV1) (Plan, []Diagnostic) {
	if diagnostics := Validate(document); len(diagnostics) > 0 {
		return Plan{}, diagnostics
	}
	encoded, _ := json.Marshal(document)
	sum := sha256.Sum256(encoded)
	plan := Plan{Revision: fmt.Sprintf("policy-v1-%x", sum[:8])}
	for _, v := range document.Corrections {
		if enabled(v.Enabled) {
			plan.Corrections = append(plan.Corrections, v)
		}
	}
	for _, v := range document.Findings {
		if enabled(v.Enabled) {
			plan.Findings = append(plan.Findings, v)
		}
	}
	for _, v := range document.Cohorts {
		if enabled(v.Enabled) {
			plan.Cohorts = append(plan.Cohorts, v)
		}
	}
	for _, v := range document.ActionRollups {
		if enabled(v.Enabled) {
			plan.ActionRollups = append(plan.ActionRollups, v)
		}
	}
	return plan, nil
}
