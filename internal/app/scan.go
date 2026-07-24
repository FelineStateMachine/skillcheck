package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"skilltrace/internal/adapters"
	"skilltrace/internal/adapters/claude"
	"skilltrace/internal/adapters/codex"
	"skilltrace/internal/adapters/huggingface"
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
	registry := adapters.NewRegistry()
	sanitizer := trace.NewHashSanitizer("skilltrace-v1")
	registry.Register("codex", codex.New(sanitizer))
	registry.Register("claude", claude.New(sanitizer))
	registry.Register("huggingface", huggingface.New(sanitizer))
	if !registry.Supports(req.Harness) {
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
	result, err := registry.Parse(ctx, req.Harness, f)
	if err != nil {
		return ScanResult{}, apperror.Wrap("malformed_source", "source could not be parsed", err)
	}
	abs, err := filepath.Abs(req.Input)
	if err != nil {
		return ScanResult{}, err
	}
	h := sha256.Sum256([]byte(filepath.Clean(abs)))
	sourceKey := req.Harness + ":" + hex.EncodeToString(h[:12])
	snapshot, err := a.Catalog.ReplaceSource(ctx, sourceKey, result)
	if err != nil {
		return ScanResult{}, apperror.Wrap("catalog_write_failed", "catalog update failed", err)
	}
	if progress != nil {
		progress(Progress{Stage: "committed", Completed: 1, Total: 1})
	}
	status := "full"
	if len(result.Exclusions) > 0 {
		status = "partial"
	}
	// Preserve the V1 Codex health label consumed by existing clients.
	if req.Harness == "codex" {
		status = "supported"
	}
	return ScanResult{Source: catalog.Health{SourceKey: sourceKey, Harness: req.Harness, Status: status, Revision: snapshot.Revision, EventCount: snapshot.EventCount, ExclusionCount: snapshot.ExclusionCount}, Capabilities: result.Capabilities}, nil
}

func (a *Application) Health() ([]catalog.Health, error) {
	h, err := a.Catalog.Health()
	if err != nil {
		return nil, fmt.Errorf("read source health: %w", err)
	}
	return h, nil
}
