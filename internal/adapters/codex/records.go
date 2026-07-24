package codex

import "encoding/json"

// record is one line of a trace. Codex rollout sessions nest their content
// under payload with a secondary type; the canonical flat contract used by
// fixtures carries its fields at the top level. Both are decoded into this and
// dispatched on Type.
type record struct {
	Type string `json:"type"`

	// Rollout form.
	Payload json.RawMessage `json:"payload"`

	// Flat canonical form.
	Model        string `json:"model"`
	Skill        string `json:"skill"`
	Tool         string `json:"tool"`
	Status       string `json:"status"`
	Operation    string `json:"operation"`
	Path         string `json:"path"`
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
}

type payloadHead struct {
	Type string `json:"type"`
}

type sessionMeta struct {
	ID            string  `json:"id"`
	CWD           string  `json:"cwd"`
	Timestamp     string  `json:"timestamp"`
	ModelProvider string  `json:"model_provider"`
	Git           gitInfo `json:"git"`
}

type gitInfo struct {
	Branch string `json:"branch"`
}

// functionCall is Codex's tool invocation; custom_tool_call shares the fields
// skilltrace needs.
type functionCall struct {
	Name   string `json:"name"`
	CallID string `json:"call_id"`
}

// functionCallOutput carries a call's result. Codex encodes success as a
// trailing "Process exited with code N" line rather than a status field.
type functionCallOutput struct {
	CallID string `json:"call_id"`
	Output string `json:"output"`
}
