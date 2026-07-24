package system_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"skilltrace/internal/app"
	"skilltrace/internal/apperror"
	"skilltrace/internal/catalog"
	"skilltrace/internal/source"
)

func TestIncrementalRefresh(t *testing.T) {
	ctx := context.Background()
	input := filepath.Join(t.TempDir(), "trace.jsonl")
	initial, err := os.ReadFile(filepath.Join("..", "..", "testdata", "traces", "codex", "minimal.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, initial, 0o600); err != nil {
		t.Fatal(err)
	}

	a, closeCatalog := newSystemApplication(t)
	defer closeCatalog()
	first, err := a.Scan(ctx, app.ScanRequest{Input: input, Harness: "codex"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	fingerprint, err := source.FingerprintFile(input, int64(len(initial)))
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := source.Checkpoint{Fingerprint: fingerprint, CommittedOffset: int64(len(initial)), ParserRevision: "codex-v1"}

	f, err := os.OpenFile(input, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"type":"usage","input_tokens":1`); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	withTrailingRecord, err := source.FingerprintFile(input, int64(len(initial)))
	if err != nil {
		t.Fatal(err)
	}
	if err := checkpoint.AppendEligible(withTrailingRecord, "codex-v1"); err != nil {
		t.Fatalf("append should remain eligible: %v", err)
	}
	if _, err := a.Scan(ctx, app.ScanRequest{Input: input, Harness: "codex"}, nil); err == nil {
		t.Fatal("incomplete trailing record unexpectedly committed")
	}
	health, err := a.Health()
	if err != nil {
		t.Fatal(err)
	}
	if len(health) != 1 || health[0].Revision != first.Source.Revision || health[0].EventCount != first.Source.EventCount {
		t.Fatalf("failed refresh changed prior snapshot: %#v", health)
	}

	f, err = os.OpenFile(input, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(",\"output_tokens\":1}\n"); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	refreshed, err := a.Scan(ctx, app.ScanRequest{Input: input, Harness: "codex"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.Source.Revision != first.Source.Revision+1 || refreshed.Source.EventCount != first.Source.EventCount+1 {
		t.Fatalf("unexpected refreshed snapshot: first=%#v refreshed=%#v", first.Source, refreshed.Source)
	}
}

func TestCatalogLifecycle(t *testing.T) {
	ctx := context.Background()
	a, closeCatalog := newSystemApplication(t)
	defer closeCatalog()
	input := filepath.Join("..", "..", "testdata", "traces", "codex", "minimal.jsonl")
	scanned, err := a.Scan(ctx, app.ScanRequest{Input: input, Harness: "codex"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	preview, err := a.PreviewLifecycle(ctx, "forget-source", scanned.Source.SourceKey)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Sources != 1 || preview.Token == "" {
		t.Fatalf("unexpected deletion preview: %#v", preview)
	}
	if err := a.ApplyLifecycle(ctx, preview); err != nil {
		t.Fatal(err)
	}
	health, err := a.CatalogStatus()
	if err != nil {
		t.Fatal(err)
	}
	if len(health) != 0 {
		t.Fatalf("forgotten source remains in catalog: %#v", health)
	}

	if _, err := a.Scan(ctx, app.ScanRequest{Input: input, Harness: "codex"}, nil); err != nil {
		t.Fatal(err)
	}
	reset, err := a.PreviewLifecycle(ctx, "reset", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.ApplyLifecycle(ctx, reset); err != nil {
		t.Fatal(err)
	}
	health, err = a.CatalogStatus()
	if err != nil {
		t.Fatal(err)
	}
	if len(health) != 0 {
		t.Fatalf("reset retained sources: %#v", health)
	}
}

func TestWriterBusy(t *testing.T) {
	ctx := context.Background()
	a, closeCatalog := newSystemApplication(t)
	defer closeCatalog()
	release, err := a.Catalog.AcquireWriter(ctx, "system-test", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	err = a.ApplyLifecycle(ctx, catalog.DeletionPreview{Action: "reset"})
	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != "writer_busy" || !errors.Is(err, catalog.ErrWriterBusy) {
		t.Fatalf("expected writer_busy application error, got %v", err)
	}
}

func newSystemApplication(t *testing.T) (*app.Application, func()) {
	t.Helper()
	c, err := catalog.Open(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	return app.New(c), func() {
		if err := c.Close(); err != nil {
			t.Errorf("close catalog: %v", err)
		}
	}
}
