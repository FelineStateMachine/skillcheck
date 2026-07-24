package codex

import "encoding/json"

type record struct {
	Type         string          `json:"type"`
	Model        string          `json:"model"`
	Skill        string          `json:"skill"`
	Tool         string          `json:"tool"`
	Status       string          `json:"status"`
	Operation    string          `json:"operation"`
	Path         string          `json:"path"`
	InputTokens  int64           `json:"input_tokens"`
	OutputTokens int64           `json:"output_tokens"`
	Payload      json.RawMessage `json:"payload"`
}
