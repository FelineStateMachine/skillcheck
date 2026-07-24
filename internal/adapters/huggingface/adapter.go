package huggingface

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"skilltrace/internal/adapters"
	"skilltrace/internal/adapters/claude"
	"skilltrace/internal/adapters/codex"
	"skilltrace/internal/trace"
	"strings"
)

type Adapter struct{ sanitizer trace.Sanitizer }

func New(s trace.Sanitizer) *Adapter { return &Adapter{sanitizer: s} }

func rawReader(raw json.RawMessage) (io.Reader, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, fmt.Errorf("missing raw trace")
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return strings.NewReader(text), nil
	}
	var records []json.RawMessage
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, fmt.Errorf("raw trace must be JSONL text or record array")
	}
	var b bytes.Buffer
	for _, record := range records {
		b.Write(record)
		b.WriteByte('\n')
	}
	return &b, nil
}

func (a *Adapter) Parse(ctx context.Context, input io.Reader) (adapters.Result, error) {
	result := adapters.Result{Harness: "huggingface"}
	s := bufio.NewScanner(input)
	s.Buffer(make([]byte, 64*1024), 16<<20)
	var line, sequence int64
	var profiles []trace.CapabilityProfile
	for s.Scan() {
		line++
		var r row
		if err := json.Unmarshal(s.Bytes(), &r); err != nil {
			return adapters.Result{}, fmt.Errorf("huggingface row %d: invalid JSON", line)
		}
		if !r.RawRetained {
			result.Eligibility = append(result.Eligibility, adapters.Eligibility{Row: line, Reason: "raw_trace_not_retained"})
			result.Exclusions = append(result.Exclusions, trace.Exclusion{Line: line, Reason: "raw_trace_not_retained"})
			continue
		}
		raw, err := rawReader(r.RawTrace)
		if err != nil {
			result.Eligibility = append(result.Eligibility, adapters.Eligibility{Row: line, Reason: "missing_raw_trace"})
			result.Exclusions = append(result.Exclusions, trace.Exclusion{Line: line, Reason: "missing_raw_trace"})
			continue
		}
		var nested adapters.Result
		switch r.Harness {
		case "codex":
			nested, err = codex.New(a.sanitizer).Parse(ctx, raw)
		case "claude":
			nested, err = claude.New(a.sanitizer).Parse(ctx, raw)
		default:
			result.Eligibility = append(result.Eligibility, adapters.Eligibility{Row: line, Reason: "unsupported_harness"})
			result.Exclusions = append(result.Exclusions, trace.Exclusion{Line: line, Reason: "unsupported_harness"})
			continue
		}
		if err != nil {
			result.Eligibility = append(result.Eligibility, adapters.Eligibility{Row: line, Reason: "malformed_raw_trace"})
			result.Exclusions = append(result.Exclusions, trace.Exclusion{Line: line, Reason: "malformed_raw_trace"})
			continue
		}
		result.Eligibility = append(result.Eligibility, adapters.Eligibility{Row: line, Eligible: true})
		profiles = append(profiles, nested.Capabilities)
		for _, event := range nested.Events {
			sequence++
			event.Sequence = sequence
			result.Events = append(result.Events, event)
		}
		result.Exclusions = append(result.Exclusions, nested.Exclusions...)
	}
	if err := s.Err(); err != nil {
		return adapters.Result{}, fmt.Errorf("read huggingface rows: %w", err)
	}
	result.Capabilities = trace.AggregateCapabilities(profiles...)
	return result, nil
}
