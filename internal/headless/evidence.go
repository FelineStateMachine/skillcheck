package headless

import "skilltrace/internal/evidence"

type EvidenceResult struct {
	Token string `json:"token"`
	State string `json:"state"`
}

func PresentEvidence(result evidence.Result) EvidenceResult {
	return EvidenceResult{Token: result.Token, State: result.State}
}
