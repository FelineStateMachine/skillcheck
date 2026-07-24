package catalog

type Model struct {
	Action, Target, PreviewToken string
	Affected                     int
	WriterBusy                   bool
}

func (m Model) Confirmable(token string) bool {
	return !m.WriterBusy && m.PreviewToken != "" && token == m.PreviewToken
}
