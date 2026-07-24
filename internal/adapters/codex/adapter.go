package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"skilltrace/internal/adapters"
	"skilltrace/internal/trace"
)

const maxRecordBytes = 4 << 20

type Adapter struct{ sanitizer trace.Sanitizer }

func New(s trace.Sanitizer) *Adapter { return &Adapter{sanitizer: s} }

func (a *Adapter) Parse(ctx context.Context, r io.Reader) (adapters.Result, error) {
	result := adapters.Result{Harness: "codex", Capabilities: trace.CodexCapabilities()}
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 64*1024), maxRecordBytes)
	var line, sequence int64
	for s.Scan() {
		line++
		if err := ctx.Err(); err != nil {
			return adapters.Result{}, err
		}
		var rec record
		if err := json.Unmarshal(s.Bytes(), &rec); err != nil {
			return adapters.Result{}, fmt.Errorf("codex record line %d: invalid JSON", line)
		}
		var payload any
		switch rec.Type {
		case "session":
			payload = trace.SessionPayload{Model: a.sanitizer.Label(rec.Model)}
		case "skill":
			payload = trace.SkillPayload{SkillToken: a.sanitizer.Token("skill", rec.Skill)}
		case "tool":
			payload = trace.ToolPayload{Tool: a.sanitizer.Label(rec.Tool), Status: a.sanitizer.Label(rec.Status)}
		case "file":
			payload = trace.FilePayload{Operation: a.sanitizer.Label(rec.Operation), PathToken: a.sanitizer.Token("path", rec.Path)}
		case "completion":
			payload = trace.CompletionPayload{Status: a.sanitizer.Label(rec.Status)}
		case "usage":
			payload = trace.UsagePayload{InputTokens: rec.InputTokens, OutputTokens: rec.OutputTokens}
		default:
			result.Exclusions = append(result.Exclusions, trace.Exclusion{Line: line, Reason: "unsupported_record_type"})
			continue
		}
		sequence++
		event, err := trace.NewEvent(rec.Type, sequence, line, payload)
		if err != nil {
			return adapters.Result{}, err
		}
		result.Events = append(result.Events, event)
	}
	if err := s.Err(); err != nil {
		return adapters.Result{}, fmt.Errorf("read codex trace: %w", err)
	}
	return result, nil
}
