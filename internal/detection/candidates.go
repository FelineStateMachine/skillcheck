package detection

import "skilltrace/internal/trace"

type Tier string

const (
	Confirmed Tier = "confirmed"
	Probable  Tier = "probable"
	Possible  Tier = "possible"
)

type Candidate struct {
	Skill    string  `json:"skill"`
	Tier     Tier    `json:"tier"`
	Actor    string  `json:"actor"`
	Start    int64   `json:"start"`
	End      int64   `json:"end"`
	Evidence []int64 `json:"evidence"`
}

func Candidates(events []trace.Event, skill string) []Candidate {
	var out []Candidate
	for _, e := range events {
		if e.Kind != "skill" {
			continue
		}
		var p trace.SkillPayload
		if trace.DecodePayload(e, &p) != nil || p.SkillToken != skill {
			continue
		}
		out = append(out, Candidate{Skill: skill, Tier: Confirmed, Actor: "primary", Start: e.Sequence, End: e.Sequence, Evidence: []int64{e.Sequence}})
	}
	return out
}
