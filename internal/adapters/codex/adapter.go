package codex

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"skilltrace/internal/adapters"
	"skilltrace/internal/trace"
)

const maxRecordBytes = 16 << 20

type Adapter struct{ sanitizer trace.Sanitizer }

func New(s trace.Sanitizer) *Adapter { return &Adapter{sanitizer: s} }

// Parse reads a Codex session.
//
// Real Codex rollouts nest content under payload: session_meta carries the
// working directory and branch, function_call carries a tool name and call_id,
// and function_call_output carries the exit result paired by call_id. The
// canonical flat contract used by fixtures carries its fields at the top level
// instead; both are handled so the same command reads real sessions and the
// pipeline's contract fixtures alike.
//
// A record that does not decode becomes an exclusion rather than aborting the
// file: one malformed line must not discard a long session.
func (a *Adapter) Parse(ctx context.Context, r io.Reader) (adapters.Result, error) {
	result := adapters.Result{Harness: "codex", Capabilities: trace.CodexCapabilities()}
	var (
		sequence  int64
		toolIndex = map[string]int{}
	)
	emit := func(line int64, kind string, payload any) {
		sequence++
		event, err := trace.NewEvent(kind, sequence, line, payload)
		if err != nil {
			result.Exclusions = append(result.Exclusions, trace.Exclusion{Line: line, Reason: "unrepresentable_event"})
			return
		}
		result.Events = append(result.Events, event)
	}

	err := adapters.ForEachLine(r, maxRecordBytes, func(line int64, data []byte, terminated bool) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		var rec record
		if err := json.Unmarshal(data, &rec); err != nil {
			if !terminated {
				return adapters.ErrIncompleteTrailingRecord
			}
			result.Exclusions = append(result.Exclusions, trace.Exclusion{Line: line, Reason: "invalid_json"})
			return nil
		}
		emitLine := func(kind string, payload any) { emit(line, kind, payload) }
		switch rec.Type {
		case "session_meta", "turn_context":
			a.captureSession(&result, rec.Payload)
		case "response_item":
			a.handleResponseItem(&result, emitLine, toolIndex, rec.Payload)
		case "event_msg":
			a.advanceEndTime(&result, rec.Payload)
		case "session", "skill", "tool", "file", "completion", "usage":
			a.handleFlat(emitLine, rec)
		case "world_state", "inter_agent_communication_metadata", "compacted":
			// Rollout context and bookkeeping records carry no tool activity.
			// They are expected, not errors, so they are ignored rather than
			// excluded — excluding them would mark every real session partial.
		default:
			result.Exclusions = append(result.Exclusions, trace.Exclusion{Line: line, Reason: "unsupported_record_type"})
		}
		return nil
	})
	if err != nil {
		return adapters.Result{}, err
	}
	return result, nil
}

func (a *Adapter) handleResponseItem(result *adapters.Result, emit func(string, any), toolIndex map[string]int, payload json.RawMessage) {
	var head payloadHead
	if json.Unmarshal(payload, &head) != nil {
		return
	}
	switch head.Type {
	case "function_call", "custom_tool_call":
		var call functionCall
		if json.Unmarshal(payload, &call) != nil || call.Name == "" {
			return
		}
		if call.CallID != "" {
			toolIndex[call.CallID] = len(result.Events)
		}
		emit("tool", trace.ToolPayload{
			Tool:   a.sanitizer.Label(call.Name),
			Status: "started",
			CallID: a.sanitizer.Token("call", call.CallID),
		})
	case "function_call_output", "custom_tool_call_output":
		var out functionCallOutput
		if json.Unmarshal(payload, &out) != nil {
			return
		}
		a.recordOutcome(result.Events, toolIndex, out)
	}
}

// recordOutcome patches the tool event a result belongs to. Codex reports
// success as a "Process exited with code N" line in the output text; anything
// other than code 0 is treated as an error.
func (a *Adapter) recordOutcome(events []trace.Event, toolIndex map[string]int, out functionCallOutput) {
	idx, ok := toolIndex[out.CallID]
	if !ok {
		return
	}
	var payload trace.ToolPayload
	if trace.DecodePayload(events[idx], &payload) != nil {
		return
	}
	payload.Status = exitStatus(out.Output)
	if updated, err := json.Marshal(payload); err == nil {
		events[idx].Payload = updated
	}
}

func exitStatus(output string) string {
	if strings.Contains(output, "Process exited with code 0") {
		return "ok"
	}
	if strings.Contains(output, "Process exited with code ") {
		return "error"
	}
	return "ok"
}

func (a *Adapter) captureSession(result *adapters.Result, payload json.RawMessage) {
	var meta sessionMeta
	if json.Unmarshal(payload, &meta) != nil {
		return
	}
	if result.Session.Key == "" && meta.ID != "" {
		result.Session.Key = meta.ID
	}
	if result.Session.Project == "" && meta.CWD != "" {
		result.Session.Project = adapters.ProjectName(meta.CWD)
	}
	if result.Session.Branch == "" && meta.Git.Branch != "" {
		result.Session.Branch = a.sanitizer.Label(meta.Git.Branch)
	}
	if meta.Timestamp != "" {
		if result.Session.StartedAt == "" {
			result.Session.StartedAt = meta.Timestamp
		}
		result.Session.EndedAt = meta.Timestamp
	}
}

func (a *Adapter) advanceEndTime(result *adapters.Result, payload json.RawMessage) {
	var ts struct {
		Timestamp string `json:"timestamp"`
	}
	if json.Unmarshal(payload, &ts) == nil && ts.Timestamp != "" {
		if result.Session.StartedAt == "" {
			result.Session.StartedAt = ts.Timestamp
		}
		result.Session.EndedAt = ts.Timestamp
	}
}

// handleFlat maps the canonical flat contract used by fixtures. It preserves
// the original one-event-per-record behaviour that the pipeline's tests pin.
func (a *Adapter) handleFlat(emit func(string, any), rec record) {
	switch rec.Type {
	case "session":
		emit("session", trace.SessionPayload{Model: a.sanitizer.Label(rec.Model)})
	case "skill":
		emit("skill", trace.SkillPayload{SkillToken: a.sanitizer.Token("skill", rec.Skill)})
	case "tool":
		emit("tool", trace.ToolPayload{Tool: a.sanitizer.Label(rec.Tool), Status: a.sanitizer.Label(rec.Status)})
	case "file":
		emit("file", trace.FilePayload{Operation: a.sanitizer.Label(rec.Operation), PathToken: a.sanitizer.Token("path", rec.Path)})
	case "completion":
		emit("completion", trace.CompletionPayload{Status: a.sanitizer.Label(rec.Status)})
	case "usage":
		emit("usage", trace.UsagePayload{InputTokens: rec.InputTokens, OutputTokens: rec.OutputTokens})
	}
}
