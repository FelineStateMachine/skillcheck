package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"skilltrace/internal/adapters"
	"skilltrace/internal/modelregistry"
	"skilltrace/internal/trace"
)

type Adapter struct{ sanitizer trace.Sanitizer }

func New(s trace.Sanitizer) *Adapter { return &Adapter{sanitizer: s} }

func (a *Adapter) Parse(ctx context.Context, input io.Reader) (adapters.Result, error) {
	result := adapters.Result{Harness: "claude", Capabilities: trace.ClaudeCapabilities()}
	s := bufio.NewScanner(input)
	s.Buffer(make([]byte, 64*1024), 4<<20)
	red := newReducer()
	var line, sequence int64
	for s.Scan() {
		line++
		if err := ctx.Err(); err != nil {
			return adapters.Result{}, err
		}
		var rec record
		if err := json.Unmarshal(s.Bytes(), &rec); err != nil {
			return adapters.Result{}, fmt.Errorf("claude record line %d: invalid JSON", line)
		}
		if rec.Version != 0 && rec.Version != 1 {
			result.Exclusions = append(result.Exclusions, trace.Exclusion{Line: line, Reason: "unsupported_record_version"})
			continue
		}
		actor := ""
		if rec.Actor != "" {
			actor = a.sanitizer.Token("actor", rec.Actor)
		}
		var kind string
		var payload any
		switch rec.Type {
		case "session", "session_start":
			id := modelregistry.Resolve(a.sanitizer.Label(rec.Model))
			kind = "session"
			payload = trace.SessionPayload{Model: id.Exact, Provider: id.Provider, Family: id.Family, Actor: actor}
		case "skill", "skill_use":
			kind = "skill"
			payload = trace.SkillPayload{SkillToken: a.sanitizer.Token("skill", rec.Skill)}
		case "tool_use":
			red.toolUse(rec.ToolUseID, rec.Tool)
			kind = "tool"
			payload = trace.ToolPayload{Tool: a.sanitizer.Label(rec.Tool), Status: "started", CallID: a.sanitizer.Token("call", rec.ToolUseID), Actor: actor}
		case "tool_result":
			tool, ok := red.toolResult(rec.ToolUseID)
			if !ok {
				result.Exclusions = append(result.Exclusions, trace.Exclusion{Line: line, Reason: "unmatched_tool_result"})
				continue
			}
			kind = "tool"
			payload = trace.ToolPayload{Tool: a.sanitizer.Label(tool), Status: a.sanitizer.Label(rec.Status), CallID: a.sanitizer.Token("call", rec.ToolUseID), Actor: actor}
		case "file":
			kind = "file"
			payload = trace.FilePayload{Operation: a.sanitizer.Label(rec.Operation), PathToken: a.sanitizer.Token("path", rec.Path)}
		case "completion":
			kind = "completion"
			payload = trace.CompletionPayload{Status: a.sanitizer.Label(rec.Status)}
		case "usage":
			kind = "usage"
			payload = trace.UsagePayload{InputTokens: rec.InputTokens, OutputTokens: rec.OutputTokens}
		case "subagent_start", "subagent_stop":
			kind = "session"
			payload = trace.SessionPayload{Actor: actor}
		default:
			result.Exclusions = append(result.Exclusions, trace.Exclusion{Line: line, Reason: "unsupported_record_type"})
			continue
		}
		sequence++
		event, err := trace.NewEvent(kind, sequence, line, payload)
		if err != nil {
			return adapters.Result{}, err
		}
		result.Events = append(result.Events, event)
	}
	if err := s.Err(); err != nil {
		return adapters.Result{}, fmt.Errorf("read claude trace: %w", err)
	}
	return result, nil
}
