package claude

import "encoding/json"

type record struct {
	Type         string          `json:"type"`
	Version      int             `json:"version"`
	Model        string          `json:"model"`
	Actor        string          `json:"actor"`
	Parent       string          `json:"parent"`
	Skill        string          `json:"skill"`
	Tool         string          `json:"tool"`
	ToolUseID    string          `json:"tool_use_id"`
	Status       string          `json:"status"`
	Operation    string          `json:"operation"`
	Path         string          `json:"path"`
	InputTokens  int64           `json:"input_tokens"`
	OutputTokens int64           `json:"output_tokens"`
	Message      json.RawMessage `json:"message"`
}
