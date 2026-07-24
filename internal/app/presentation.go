package app

import (
	"sort"
	"time"

	"skilltrace/internal/catalog"
	"skilltrace/internal/detection"
	"skilltrace/internal/skills"
)

type SourceSnapshot struct {
	Harness        string
	Status         string
	Revision       int64
	EventCount     int
	ExclusionCount int
	Freshness      string
}

type SkillSnapshot struct {
	ID          string
	Name        string
	Description string
	Installed   bool
	Exposures   []skills.Exposure
	State       string
	Confirmed   int
	Probable    int
	Possible    int
}

type EpisodeSnapshot struct {
	Actor        string
	Outcome      string
	Tier         detection.Tier
	Capabilities []string
	Evidence     []string
}

type DiscoverySnapshot struct {
	Skills    []SkillSnapshot
	Sources   []SourceSnapshot
	Issues    []skills.Issue
	UpdatedAt time.Time
}

func PresentDiscovery(discovered DiscoverResult, health []catalog.Health, now time.Time) DiscoverySnapshot {
	snapshot := DiscoverySnapshot{Issues: append([]skills.Issue(nil), discovered.Issues...), UpdatedAt: now}
	for _, skill := range discovered.Skills {
		snapshot.Skills = append(snapshot.Skills, SkillSnapshot{
			ID: skill.Identity.ID, Name: skill.Identity.Name, Description: skill.Identity.Description,
			Installed: true, Exposures: append([]skills.Exposure(nil), skill.Exposures...), State: "cached",
		})
	}
	for _, source := range health {
		snapshot.Sources = append(snapshot.Sources, SourceSnapshot{
			Harness: source.Harness, Status: source.Status, Revision: source.Revision,
			EventCount: source.EventCount, ExclusionCount: source.ExclusionCount, Freshness: "cached",
		})
	}
	sort.Slice(snapshot.Skills, func(i, j int) bool { return snapshot.Skills[i].Name < snapshot.Skills[j].Name })
	return snapshot
}

func PresentEpisodes(result AnalyzeResult) []EpisodeSnapshot {
	episodes := make([]EpisodeSnapshot, 0, len(result.Episodes))
	for _, episode := range result.Episodes {
		evidence := make([]string, len(episode.EventSequences))
		for i, sequence := range episode.EventSequences {
			evidence[i] = evidenceToken(sequence)
		}
		episodes = append(episodes, EpisodeSnapshot{
			Actor: episode.Actor, Outcome: episode.Outcome, Tier: episode.Tier,
			Capabilities: []string{"actor", "outcome", "evidence"}, Evidence: evidence,
		})
	}
	return episodes
}

func evidenceToken(sequence int64) string {
	const alphabet = "0123456789abcdef"
	if sequence == 0 {
		return "event-0"
	}
	b := make([]byte, 0, 16)
	for sequence > 0 {
		b = append(b, alphabet[sequence&15])
		sequence >>= 4
	}
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return "event-" + string(b)
}
