package app

import (
	"crypto/sha256"
	"fmt"
	"os"

	"skilltrace/internal/catalog"
	"skilltrace/internal/policy"
)

type PolicyValidation struct {
	Valid       bool                `json:"valid"`
	Fingerprint string              `json:"fingerprint,omitempty"`
	Revision    string              `json:"revision,omitempty"`
	Diagnostics []policy.Diagnostic `json:"diagnostics,omitempty"`
}
type PolicyPreview struct {
	Token              string        `json:"token"`
	Revision           string        `json:"revision"`
	Fingerprint        string        `json:"fingerprint"`
	Corrections        int           `json:"corrections"`
	Findings           int           `json:"findings"`
	Cohorts            int           `json:"cohorts"`
	ActionRollups      int           `json:"action_rollups"`
	AffectedCandidates int           `json:"affected_candidates"`
	Invalidates        Invalidations `json:"invalidates"`
}
type PolicyApplyResult struct {
	Applied bool `json:"applied"`
	PolicyPreview
}

func (a *Application) ValidatePolicy(path string) (PolicyValidation, error) {
	doc, diagnostics, err := (policy.Store{}).Read(path)
	if err != nil {
		return PolicyValidation{}, err
	}
	if len(diagnostics) > 0 {
		return PolicyValidation{Valid: false, Diagnostics: diagnostics}, nil
	}
	plan, diagnostics := policy.Compile(doc.V1)
	if len(diagnostics) > 0 {
		return PolicyValidation{Valid: false, Diagnostics: diagnostics}, nil
	}
	return PolicyValidation{Valid: true, Fingerprint: doc.Fingerprint, Revision: plan.Revision}, nil
}

func (a *Application) PreviewPolicy(path string) (PolicyPreview, error) {
	doc, diagnostics, err := (policy.Store{}).Read(path)
	if err != nil {
		return PolicyPreview{}, err
	}
	if len(diagnostics) > 0 {
		return PolicyPreview{}, diagnostics[0]
	}
	plan, diagnostics := policy.Compile(doc.V1)
	if len(diagnostics) > 0 {
		return PolicyPreview{}, diagnostics[0]
	}
	affected := 0
	for _, correction := range plan.Corrections {
		filter := policy.Pushdown(correction.Match)
		count, err := a.Catalog.CountPolicyCandidates(filter.Clause, filter.Args...)
		if err != nil {
			return PolicyPreview{}, err
		}
		affected += count
	}
	sum := sha256.Sum256([]byte(plan.Revision + ":" + doc.Fingerprint + fmt.Sprint(affected)))
	return PolicyPreview{Token: fmt.Sprintf("preview-%x", sum[:12]), Revision: plan.Revision, Fingerprint: doc.Fingerprint, Corrections: len(plan.Corrections), Findings: len(plan.Findings), Cohorts: len(plan.Cohorts), ActionRollups: len(plan.ActionRollups), AffectedCandidates: affected, Invalidates: policyInvalidations(len(plan.Corrections), len(plan.Findings), len(plan.Cohorts), len(plan.ActionRollups))}, nil
}

func (a *Application) ApplyPolicy(path, token string) (PolicyApplyResult, error) {
	preview, err := a.PreviewPolicy(path)
	if err != nil {
		return PolicyApplyResult{}, err
	}
	if token == "" || token != preview.Token {
		return PolicyApplyResult{}, policy.ConflictError{Message: "policy preview is stale or missing"}
	}
	if _, err := os.Stat(path); err != nil {
		return PolicyApplyResult{}, err
	}
	err = a.Catalog.ApplyPolicy(catalog.PolicyAudit{Revision: preview.Revision, PreviewToken: preview.Token, AffectedCorrections: preview.AffectedCandidates, AffectedFindings: preview.Findings}, preview.Fingerprint, "policy-file")
	if err != nil {
		return PolicyApplyResult{}, err
	}
	return PolicyApplyResult{Applied: true, PolicyPreview: preview}, nil
}
