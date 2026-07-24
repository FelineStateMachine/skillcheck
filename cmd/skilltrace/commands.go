package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"skilltrace/internal/app"
	"skilltrace/internal/apperror"
	"skilltrace/internal/catalog"
	"skilltrace/internal/headless"
)

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: skilltrace <source|discover|analyze> [options]")
		return 2
	}
	if args[0] == "source" {
		return runSource(ctx, args, stdout, stderr)
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)
	format := fs.String("format", "text", "output format: text or json")
	catalogPath := fs.String("catalog", "", "catalog path")
	skillRoot := fs.String("skill-root", "", "skill installation root")
	skill := fs.String("skill", "", "skill name")
	scope := fs.String("scope", "current", "analysis scope")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if *catalogPath == "" {
		*catalogPath = defaultCatalogPath()
	}
	c, err := catalog.Open(*catalogPath)
	if err != nil {
		return renderError(stdout, *format, args[1], apperror.Wrap("catalog_unavailable", "catalog is unavailable", err))
	}
	defer c.Close()
	a := app.New(c)
	switch args[0] {
	case "discover":
		if *skillRoot == "" {
			*skillRoot = filepath.Join(".codex", "skills")
		}
		result, err := a.Discover(app.DiscoverRequest{SkillRoot: *skillRoot})
		if err != nil {
			return renderError(stdout, *format, "discover", err)
		}
		if err := headless.Render(stdout, *format, headless.Success("discover", result)); err != nil {
			return 1
		}
		return 0
	case "analyze":
		if *skill == "" {
			return renderError(stdout, *format, "analyze", apperror.Wrap("invalid_arguments", "skill is required", nil))
		}
		result, err := a.Analyze(ctx, app.AnalyzeRequest{Skill: *skill, Scope: *scope})
		if err != nil {
			return renderError(stdout, *format, "analyze", err)
		}
		if err := headless.Render(stdout, *format, headless.Success("analyze", result)); err != nil {
			return 1
		}
		return 0
	default:
		fmt.Fprintln(stderr, "unknown command")
		return 2
	}
}

func runSource(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: skilltrace source <scan|health> [options]")
		return 2
	}
	fs := flag.NewFlagSet("source "+args[1], flag.ContinueOnError)
	fs.SetOutput(stderr)
	format := fs.String("format", "text", "output format")
	catalogPath := fs.String("catalog", "", "catalog path")
	input := fs.String("input", "", "trace JSONL path")
	harness := fs.String("harness", "", "trace harness")
	if err := fs.Parse(args[2:]); err != nil {
		return 2
	}
	if *catalogPath == "" {
		*catalogPath = defaultCatalogPath()
	}
	c, err := catalog.Open(*catalogPath)
	if err != nil {
		return renderError(stdout, *format, args[1], apperror.Wrap("catalog_unavailable", "catalog is unavailable", err))
	}
	defer c.Close()
	a := app.New(c)
	switch args[1] {
	case "scan":
		if *input == "" || *harness == "" {
			return renderError(stdout, *format, "source scan", apperror.Wrap("invalid_arguments", "input and harness are required", nil))
		}
		result, err := a.Scan(ctx, app.ScanRequest{Input: *input, Harness: *harness}, nil)
		if err != nil {
			return renderError(stdout, *format, "source scan", err)
		}
		if err := headless.Render(stdout, *format, headless.Success("source scan", result)); err != nil {
			fmt.Fprintln(stderr, "render failed")
			return 1
		}
		return 0
	case "health":
		result, err := a.Health()
		if err != nil {
			return renderError(stdout, *format, "source health", err)
		}
		if err := headless.Render(stdout, *format, headless.Success("source health", result)); err != nil {
			return 1
		}
		return 0
	default:
		fmt.Fprintln(stderr, "unknown source command")
		return 2
	}
}

func renderError(w io.Writer, format, command string, err error) int {
	var ae *apperror.Error
	code, message := "internal_error", "operation failed"
	if errors.As(err, &ae) {
		code, message = ae.Code, ae.Message
	}
	_ = headless.Render(w, format, headless.Failure(command, code, message))
	return 1
}

func defaultCatalogPath() string {
	if p := os.Getenv("SKILLTRACE_CATALOG"); p != "" {
		return p
	}
	d, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "skilltrace", "catalog.db")
	}
	return filepath.Join(d, "skilltrace", "catalog.db")
}
