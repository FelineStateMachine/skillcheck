package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"skilltrace/internal/adapters/codex"
	"skilltrace/internal/apperror"
	"skilltrace/internal/catalog"
	"skilltrace/internal/trace"
)

type ScanRequest struct{ Input, Harness string }
type ScanResult struct {
	Source       catalog.Health          `json:"source"`
	Capabilities trace.CapabilityProfile `json:"capabilities"`
}

func (a *Application) Scan(ctx context.Context, req ScanRequest, progress ProgressFunc) (ScanResult, error) {
	if req.Harness != "codex" {
		return ScanResult{}, apperror.Wrap("unsupported_harness", "unsupported harness", nil)
	}
	f, err := os.Open(req.Input)
	if err != nil {
		return ScanResult{}, apperror.Wrap("source_unavailable", "source is unavailable", err)
	}
	defer f.Close()
	if progress != nil {
		progress(Progress{Stage: "parsing", Completed: 0, Total: 1})
	}
	result, err := codex.New(trace.NewHashSanitizer("skilltrace-v1")).Parse(ctx, f)
	if err != nil {
		return ScanResult{}, apperror.Wrap("malformed_source", "source could not be parsed", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return ScanResult{}, err
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ScanResult{}, err
	}
	sourceKey := req.Harness + ":" + hex.EncodeToString(h.Sum(nil)[:12])
	snapshot, err := a.Catalog.ReplaceSource(ctx, sourceKey, result)
	if err != nil {
		return ScanResult{}, apperror.Wrap("catalog_write_failed", "catalog update failed", err)
	}
	if progress != nil {
		progress(Progress{Stage: "committed", Completed: 1, Total: 1})
	}
	return ScanResult{Source: catalog.Health{SourceKey: sourceKey, Harness: req.Harness, Status: "supported", Revision: snapshot.Revision, EventCount: snapshot.EventCount, ExclusionCount: snapshot.ExclusionCount}, Capabilities: result.Capabilities}, nil
}

func (a *Application) Health() ([]catalog.Health, error) {
	h, err := a.Catalog.Health()
	if err != nil {
		return nil, fmt.Errorf("read source health: %w", err)
	}
	return h, nil
}
