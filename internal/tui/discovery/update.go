package discovery

type MoveMsg int
type ResizeMsg struct{ Width, Height int }
type ScanStartedMsg struct{}
type ScanProgressMsg struct {
	Stage       string
	Done, Total int
}
type ScanFinishedMsg struct{ Err error }

func (m Model) Update(msg any) Model {
	switch msg := msg.(type) {
	case MoveMsg:
		m.Cursor += int(msg)
		if m.Cursor < 0 {
			m.Cursor = 0
		}
		if last := len(m.Snapshot.Skills) - 1; last >= 0 && m.Cursor > last {
			m.Cursor = last
		}
	case ResizeMsg:
		m.Width, m.Height = msg.Width, msg.Height
	case ScanStartedMsg:
		m.Scanning, m.Error = true, ""
	case ScanProgressMsg:
		m.Scanning, m.Stage, m.Done, m.Total = true, msg.Stage, msg.Done, msg.Total
	case ScanFinishedMsg:
		m.Scanning = false
		if msg.Err != nil {
			m.Error = msg.Err.Error()
		}
	}
	return m
}
