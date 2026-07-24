package claude

type reducer struct{ calls map[string]string }

func newReducer() *reducer { return &reducer{calls: map[string]string{}} }
func (r *reducer) toolUse(id, tool string) {
	if id != "" {
		r.calls[id] = tool
	}
}
func (r *reducer) toolResult(id string) (string, bool) {
	tool, ok := r.calls[id]
	if ok {
		delete(r.calls, id)
	}
	return tool, ok
}
