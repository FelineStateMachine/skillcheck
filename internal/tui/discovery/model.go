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
	// Analyzing marks an in-flight analyze so the view can say so; without it
	// a multi-second analyze is indistinguishable from a dropped keystroke.
	Analyzing bool
	// AnalyzingSkill names the skill being analyzed.
	AnalyzingSkill string
	// Roots are the skill directories that were searched, shown when nothing
	// was found so an empty list is diagnosable.
	Roots []string
	// ScanAvailable reports whether a scan input is configured; the footer
	// only advertises the key when pressing it would do something.
	ScanAvailable bool
}

func New(snapshot app.DiscoverySnapshot) Model { return Model{Snapshot: snapshot} }

func (m Model) Selected() (app.SkillSnapshot, bool) {
	if m.Cursor < 0 || m.Cursor >= len(m.Snapshot.Skills) {
		return app.SkillSnapshot{}, false
	}
	return m.Snapshot.Skills[m.Cursor], true
}

// Busy reports whether an operation is running, so the root can ignore repeat
// activations rather than cancelling and restarting the work.
func (m Model) Busy() bool { return m.Scanning || m.Analyzing }
