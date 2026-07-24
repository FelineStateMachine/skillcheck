package claude

import (
	"context"
	"encoding/json"
	"io"

	"skilltrace/internal/adapters"
	"skilltrace/internal/trace"
)

const maxRecordBytes = 16 << 20

type Adapter struct{ sanitizer trace.Sanitizer }

func New(s trace.Sanitizer) *Adapter { return &Adapter{sanitizer: s} }

// Parse reads a real Claude Code transcript.
//
// The transcript is a stream of assistant/user/system records. Tool use lives
// inside assistant messages as tool_use content blocks, and outcomes arrive
// later as tool_result blocks in user messages, paired by tool_use_id. The
// harness also stamps each assistant turn with attributionSkill — the skill it
// considers active — which is a far stronger skill signal than the Skill tool
// firing, and the one that closes the "referenced but never invoked" gap.
//
// Records that do not decode, or that carry a type skilltrace does not model,
// become exclusions or are ignored rather than aborting the file: one
// malformed line in a thousand-line transcript must not discard the session.
func (a *Adapter) Parse(ctx context.Context, input io.Reader) (adapters.Result, error) {
	result := adapters.Result{Harness: "claude", Capabilities: trace.ClaudeCapabilities()}
	var (
		sequence    int64
		activeSkill string
		toolIndex   = map[string]int{} // tool_use_id -> index into result.Events
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

	err := adapters.ForEachLine(input, maxRecordBytes, func(line int64, data []byte, terminated bool) error {
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

		a.captureSession(&result, rec)
		actor := actorLabel(rec)

		switch rec.Type {
		case "assistant":
			// The skill that owns this turn. Emitting only on change turns a
			// long run of turns under one skill into a single episode span
			// rather than one episode per turn.
			if skill := a.skillToken(rec); skill != "" && skill != activeSkill {
				activeSkill = skill
				emit(line, "skill", trace.SkillPayload{SkillToken: skill})
			}
			for _, b := range contentBlocks(rec.Message.Content) {
				if b.Type != "tool_use" || b.Name == "" {
					continue
				}
				toolIndex[b.ID] = len(result.Events)
				emit(line, "tool", trace.ToolPayload{
					Tool:   a.sanitizer.Label(b.Name),
					Status: "started",
					CallID: a.sanitizer.Token("call", b.ID),
					Actor:  actor,
				})
			}
		case "user":
			for _, b := range contentBlocks(rec.Message.Content) {
				if b.Type == "tool_result" {
					a.recordOutcome(result.Events, toolIndex, b)
				}
			}
		default:
			// mode, system, attachment, file-history-snapshot and the rest
			// carry no tool activity; they are context, not exclusions.
		}
		return nil
	})
	if err != nil {
		return adapters.Result{}, err
	}
	return result, nil
}

// captureSession fills session metadata from the first record that carries it.
// Project is reduced to a bare directory name by adapters.ProjectName.
func (a *Adapter) captureSession(result *adapters.Result, rec record) {
	if result.Session.Key == "" && rec.SessionID != "" {
		result.Session.Key = rec.SessionID
	}
	if result.Session.Project == "" && rec.CWD != "" {
		result.Session.Project = adapters.ProjectName(rec.CWD)
	}
	if result.Session.Branch == "" && rec.GitBranch != "" {
		result.Session.Branch = a.sanitizer.Label(rec.GitBranch)
	}
	if rec.Timestamp != "" {
		if result.Session.StartedAt == "" {
			result.Session.StartedAt = rec.Timestamp
		}
		result.Session.EndedAt = rec.Timestamp
	}
}

// skillToken resolves the skill active on an assistant turn, preferring the
// harness's own attribution and falling back to an explicit Skill invocation.
func (a *Adapter) skillToken(rec record) string {
	if rec.AttributionSkill != "" {
		return a.sanitizer.Token("skill", rec.AttributionSkill)
	}
	for _, b := range contentBlocks(rec.Message.Content) {
		if b.Type == "tool_use" && b.Name == "Skill" {
			var in skillInput
			if json.Unmarshal(b.Input, &in) == nil && in.Skill != "" {
				return a.sanitizer.Token("skill", in.Skill)
			}
		}
	}
	return ""
}

// recordOutcome patches the tool event a result belongs to, so each tool call
// is one event carrying its own outcome instead of a started/finished pair.
func (a *Adapter) recordOutcome(events []trace.Event, toolIndex map[string]int, b block) {
	idx, ok := toolIndex[b.ToolUseID]
	if !ok {
		return
	}
	var payload trace.ToolPayload
	if trace.DecodePayload(events[idx], &payload) != nil {
		return
	}
	payload.Status = "ok"
	if b.IsError {
		payload.Status = "error"
	}
	if updated, err := json.Marshal(payload); err == nil {
		events[idx].Payload = updated
	}
}

func actorLabel(rec record) string {
	if rec.IsSidechain {
		return "subagent"
	}
	return "primary"
}

// contentBlocks returns the block array of a message, tolerating the plain
// string form that user messages sometimes use instead of an array.
func contentBlocks(raw json.RawMessage) []block {
	if len(raw) == 0 {
		return nil
	}
	var blocks []block
	if json.Unmarshal(raw, &blocks) == nil {
		return blocks
	}
	return nil
}
