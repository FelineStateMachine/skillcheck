package detection

import (
	"skilltrace/internal/trace"
	"testing"
)

func TestPassesBoundEpisodesAndClassifyOutcome(t *testing.T) {
	events := []trace.Event{}
	values := []struct {
		kind  string
		value any
	}{{"skill", trace.SkillPayload{SkillToken: "skill:a"}}, {"tool", trace.ToolPayload{Tool: "shell"}}, {"completion", trace.CompletionPayload{Status: "success"}}}
	for i, p := range values {
		e, err := trace.NewEvent(p.kind, int64(i+1), int64(i+1), p.value)
		if err != nil {
			t.Fatal(err)
		}
		events = append(events, e)
	}
	episodes := Attribute(events, Candidates(events, "skill:a"))
	if len(episodes) != 1 || episodes[0].Outcome != "succeeded" || len(episodes[0].EventSequences) != 3 {
		t.Fatalf("unexpected episodes: %#v", episodes)
	}
}
