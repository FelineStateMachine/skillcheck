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
}

type Adapter interface {
	Parse(context.Context, io.Reader) (Result, error)
}
