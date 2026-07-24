package main

import (
	"context"
	"flag"
	"io"

	"skilltrace/internal/app"
	"skilltrace/internal/apperror"
	"skilltrace/internal/catalog"
	"skilltrace/internal/headless"
)

// runSync walks the machine's trace directories and scans everything new or
// changed. With no --claude-root/--codex-root it uses the default locations, so
// a bare `skilltrace sync` brings the whole machine into the catalog.
func runSync(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(stderr)
	format := fs.String("format", "text", "output format: text or json")
	catalogPath := fs.String("catalog", "", "catalog path")
	claudeRoot := fs.String("claude-root", "", "override the Claude transcript directory")
	codexRoot := fs.String("codex-root", "", "override the Codex sessions directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *catalogPath == "" {
		*catalogPath = defaultCatalogPath()
	}
	c, err := catalog.Open(*catalogPath)
	if err != nil {
		return renderError(stdout, *format, "sync", apperror.Wrap("catalog_unavailable", "catalog is unavailable", err))
	}
	defer c.Close()

	var roots []app.SyncRoot
	if *claudeRoot != "" {
		roots = append(roots, app.SyncRoot{Harness: "claude", Path: *claudeRoot})
	}
	if *codexRoot != "" {
		roots = append(roots, app.SyncRoot{Harness: "codex", Path: *codexRoot})
	}

	result, err := app.New(c).Sync(ctx, app.SyncRequest{Roots: roots}, func(scanned, skipped, failed int) {
		_ = headless.RenderProgress(stderr, "sync", "scanning", scanned+skipped+failed, 0)
	})
	if err != nil {
		return renderError(stdout, *format, "sync", err)
	}
	if err := headless.Render(stdout, *format, headless.Success("sync", result)); err != nil {
		return 1
	}
	return 0
}
