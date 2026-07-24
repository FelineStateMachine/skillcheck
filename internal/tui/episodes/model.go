package episodes

import "skilltrace/internal/app"

type Model struct {
	Skill    string
	Episodes []app.EpisodeSnapshot
	Cursor   int
	Width    int
	Height   int
}

func New(skill string, episodes []app.EpisodeSnapshot) Model {
	return Model{Skill: skill, Episodes: append([]app.EpisodeSnapshot(nil), episodes...)}
}
