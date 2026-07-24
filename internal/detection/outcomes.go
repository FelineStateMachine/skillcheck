package detection

import "skilltrace/internal/trace"

func Outcome(events []trace.Event, start, end int64) string {
	for _, e := range events {
		if e.Sequence < start || e.Sequence > end || e.Kind != "completion" {
			continue
		}
		if outcome := completionOutcome(e); outcome != "" {
			return outcome
		}
	}
	return "unknown"
}

// completionOutcome maps a completion event to an episode outcome, returning
// the empty string when the payload is undecodable or carries a status that
// does not settle the episode. Callers keep scanning on an empty result.
func completionOutcome(event trace.Event) string {
	var p trace.CompletionPayload
	if trace.DecodePayload(event, &p) != nil {
		return ""
	}
	switch p.Status {
	case "success", "completed":
		return "succeeded"
	case "failure", "failed":
		return "failed"
	}
	return ""
}
