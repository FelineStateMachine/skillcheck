package trace

type SessionPayload struct {
	Model string `json:"model,omitempty"`
}
type SkillPayload struct {
	SkillToken string `json:"skill_token"`
}
type ToolPayload struct {
	Tool   string `json:"tool"`
	Status string `json:"status,omitempty"`
}
type FilePayload struct {
	Operation string `json:"operation"`
	PathToken string `json:"path_token"`
}
type CompletionPayload struct {
	Status string `json:"status"`
}
type UsagePayload struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}

type Exclusion struct {
	Line   int64  `json:"line"`
	Reason string `json:"reason"`
}
