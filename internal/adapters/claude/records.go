package claude

import "encoding/json"

// record is one line of a Claude Code transcript. Only the fields skilltrace
// consumes are declared; the format carries much more.
//
// Version is deliberately json.RawMessage: real transcripts carry a semver
// string ("2.1.205"), while the original fixtures used an integer. Decoding it
// into an int made every real line fail as "invalid JSON" when the JSON was in
// fact valid. It is not used for gating.
type record struct {
	Type             string          `json:"type"`
	Version          json.RawMessage `json:"version"`
	SessionID        string          `json:"sessionId"`
	CWD              string          `json:"cwd"`
	GitBranch        string          `json:"gitBranch"`
	Timestamp        string          `json:"timestamp"`
	IsSidechain      bool            `json:"isSidechain"`
	AttributionSkill string          `json:"attributionSkill"`
	Message          message         `json:"message"`
}

type message struct {
	Role    string          `json:"role"`
	Model   string          `json:"model"`
	Content json.RawMessage `json:"content"`
}

// block is one content block inside an assistant or user message. Content is
// sometimes a plain string and sometimes an array of these; only tool_use and
// tool_result carry anything skilltrace records.
type block struct {
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	ID        string          `json:"id"`
	ToolUseID string          `json:"tool_use_id"`
	IsError   bool            `json:"is_error"`
	Input     json.RawMessage `json:"input"`
}

type skillInput struct {
	Skill string `json:"skill"`
}
