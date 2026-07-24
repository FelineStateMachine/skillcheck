package adapters

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"skilltrace/internal/trace"
)

type Result struct {
	Harness      string
	Events       []trace.Event
	Exclusions   []trace.Exclusion
	Capabilities trace.CapabilityProfile
	Eligibility  []Eligibility
	Session      SessionMeta
}

// SessionMeta describes the run a trace came from.
//
// Project is the bare directory name of the working directory, never the path
// that contains it: docs/privacy.md keeps source paths out of stored records,
// and a folder name is enough to group and label work.
type SessionMeta struct {
	Key       string
	Project   string
	Branch    string
	StartedAt string
	EndedAt   string
}

// ProjectName reduces a working directory to the label stored for it, so
// "/Users/dami/Developer/lofi" is recorded as "lofi".
func ProjectName(workingDir string) string {
	cleaned := strings.TrimRight(filepath.Clean(strings.TrimSpace(workingDir)), string(filepath.Separator))
	if cleaned == "" || cleaned == "." || cleaned == string(filepath.Separator) {
		return ""
	}
	base := filepath.Base(cleaned)
	if base == "." || base == string(filepath.Separator) {
		return ""
	}
	return base
}

type Eligibility struct {
	Row      int64  `json:"row"`
	Eligible bool   `json:"eligible"`
	Reason   string `json:"reason,omitempty"`
}

type Adapter interface {
	Parse(context.Context, io.Reader) (Result, error)
}
