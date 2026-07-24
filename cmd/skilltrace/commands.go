package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/mattn/go-isatty"
	"skilltrace/internal/app"
	"skilltrace/internal/apperror"
	"skilltrace/internal/catalog"
	"skilltrace/internal/headless"
	"skilltrace/internal/tui"
)

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		return runTUI(stdout, stderr)
	}
	if args[0] == "source" {
		return runSource(ctx, args, stdout, stderr)
	}
	if args[0] == "benchmark" {
		return runBenchmark(ctx, args, stdout, stderr)
	}
	if args[0] == "compare" {
		return runCompare(ctx, args[1:], stdout, stderr)
	}
	if args[0] == "policy" {
		return runPolicy(args[1:], stdout, stderr)
	}
	if args[0] == "correction" {
		return runCorrection(args[1:], stdout, stderr)
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

func runTUI(stdout, stderr io.Writer) int {
	c, err := catalog.Open(defaultCatalogPath())
	if err != nil {
		fmt.Fprintln(stderr, "catalog is unavailable")
		return 1
	}
	defer c.Close()
	skillRoot := os.Getenv("SKILLTRACE_SKILL_ROOT")
	if skillRoot == "" {
		skillRoot = filepath.Join(".codex", "skills")
	}
	color := false
	if f, ok := stdout.(*os.File); ok {
		color = isatty.IsTerminal(f.Fd()) && os.Getenv("TERM") != "dumb" && os.Getenv("NO_COLOR") == ""
	}
	locale := strings.ToUpper(os.Getenv("LC_ALL") + os.Getenv("LC_CTYPE") + os.Getenv("LANG"))
	root, err := tui.Load(app.New(c), tui.Config{
		SkillRoot: skillRoot, ScanInput: os.Getenv("SKILLTRACE_SCAN_INPUT"), Harness: "codex",
		Color: color, Unicode: strings.Contains(locale, "UTF-8") || strings.Contains(locale, "UTF8"),
	})
	if err != nil {
		fmt.Fprintln(stderr, "could not load discovery")
		return 1
	}
	if _, err := tea.NewProgram(root, tea.WithOutput(stdout)).Run(); err != nil {
		fmt.Fprintln(stderr, "terminal session failed")
		return 1
	}
	return 0
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
