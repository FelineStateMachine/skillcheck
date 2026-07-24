package tui

import tea "charm.land/bubbletea/v2"

type OperationStartedMsg struct{ ID uint64 }
type OperationProgressMsg struct {
	ID               uint64
	Stage            string
	Completed, Total int
}
type OperationFinishedMsg struct {
	ID     uint64
	Result any
	Err    error
}

func Await[T any](id uint64, events <-chan T) tea.Cmd {
	return func() tea.Msg {
		value, ok := <-events
		if !ok {
			return OperationFinishedMsg{ID: id}
		}
		return value
	}
}
