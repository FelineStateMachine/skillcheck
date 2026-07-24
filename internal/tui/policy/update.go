package policy

type PreviewLoaded struct{ Preview any }
type ApplyCompleted struct{}
type Failed struct{ Message string }

func (m Model) ApplyCompleted() Model     { m.Applied = true; m.Err = ""; return m }
func (m Model) Fail(message string) Model { m.Err = message; return m }
