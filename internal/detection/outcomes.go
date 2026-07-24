package detection

import "skilltrace/internal/trace"

func Outcome(events []trace.Event, start, end int64) string {
	for _, e := range events {
		if e.Sequence < start || e.Sequence > end || e.Kind != "completion" {
			continue
		}
		var p trace.CompletionPayload
		if trace.DecodePayload(e, &p) == nil {
			if p.Status == "success" || p.Status == "completed" {
				return "succeeded"
			}
			if p.Status == "failure" || p.Status == "failed" {
				return "failed"
			}
		}
	}
	return "unknown"
}
