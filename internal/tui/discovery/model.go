package discovery

import "skilltrace/internal/app"

type Model struct {
	Snapshot app.DiscoverySnapshot
	Cursor   int
	Width    int
	Height   int
	Scanning bool
	Stage    string
	Done     int
	Total    int
	Error    string
}

func New(snapshot app.DiscoverySnapshot) Model { return Model{Snapshot: snapshot} }

func (m Model) Selected() (app.SkillSnapshot, bool) {
	if m.Cursor < 0 || m.Cursor >= len(m.Snapshot.Skills) {
		return app.SkillSnapshot{}, false
	}
	return m.Snapshot.Skills[m.Cursor], true
}
