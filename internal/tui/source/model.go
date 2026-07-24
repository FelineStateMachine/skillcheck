package source

type Model struct {
	Harness, Status string
	Stale           bool
	Revision        int64
}

func (m Model) Label() string {
	if m.Stale {
		return m.Harness + ": stale (last good snapshot)"
	}
	return m.Harness + ": " + m.Status
}
