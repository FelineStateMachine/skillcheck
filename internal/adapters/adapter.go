package adapters

import (
	"context"
	"io"

	"skilltrace/internal/trace"
)

type Result struct {
	Harness      string
	Events       []trace.Event
	Exclusions   []trace.Exclusion
	Capabilities trace.CapabilityProfile
	Eligibility  []Eligibility
}

type Eligibility struct {
	Row      int64  `json:"row"`
	Eligible bool   `json:"eligible"`
	Reason   string `json:"reason,omitempty"`
}

type Adapter interface {
	Parse(context.Context, io.Reader) (Result, error)
}
