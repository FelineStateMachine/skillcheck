package policy

import (
	"skilltrace/internal/detection"
	"skilltrace/internal/trace"
	"skilltrace/internal/workflow"
)

type CandidateInput struct {
	Skill, Actor, Source, Model string
	Automated                   detection.Tier
}

type CorrectionResult struct {
	Automated      detection.Tier `json:"automated"`
	Final          string         `json:"final"`
	CorrectionID   string         `json:"correction_id,omitempty"`
	PolicyRevision string         `json:"policy_revision,omitempty"`
}

func (p Plan) Correct(in CandidateInput) CorrectionResult {
	result := CorrectionResult{Automated: in.Automated, Final: string(in.Automated)}
	for _, rule := range p.Corrections {
		if matches(rule.Match, in, string(in.Automated)) {
			result.Final, result.CorrectionID, result.PolicyRevision = rule.Label, rule.ID, p.Revision
		}
	}
	return result
}

type FindingInput struct {
	Candidate    CandidateInput
	Samples      int
	Capabilities trace.CapabilityProfile
}

func (p Plan) EvaluateFindings(in FindingInput) []workflow.Finding {
	var result []workflow.Finding
	for _, rule := range p.Findings {
		if in.Samples < rule.MinSamples || !matches(rule.Match, in.Candidate, string(in.Candidate.Automated)) || !hasCapabilities(in.Capabilities, rule.Capabilities) {
			continue
		}
		result = append(result, workflow.Finding{ID: rule.ID, Kind: rule.Kind, Title: rule.Title, Detail: rule.Detail})
	}
	return result
}

func matches(m Match, in CandidateInput, tier string) bool {
	return (m.Skill == "" || m.Skill == in.Skill) && (m.Actor == "" || m.Actor == in.Actor) &&
		(m.Source == "" || m.Source == in.Source) && (m.Model == "" || m.Model == in.Model) && (m.Tier == "" || m.Tier == tier)
}

func hasCapabilities(c trace.CapabilityProfile, required []string) bool {
	for _, name := range required {
		if (name == "tool_calls" && !c.ToolCalls) || (name == "usage" && !c.Usage) || (name == "actors" && !c.Actors) || (name == "skill_evidence" && !c.SkillEvidence) || (name == "file_events" && !c.FileEvents) {
			return false
		}
	}
	return true
}
