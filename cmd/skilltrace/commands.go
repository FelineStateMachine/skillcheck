package main

import (
	"context"
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
	if len(args) < 1 || strings.HasPrefix(args[0], "-") {
		return runTUI(args, stdout, stderr)
	}
	if args[0] == "source" {
		return runSource(ctx, args, stdout, stderr)
	}
	if args[0] == "sync" {
		return runSync(ctx, args[1:], stdout, stderr)
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
	if args[0] == "export" {
		return runExport(ctx, args[1:], stdout, stderr)
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
		// An empty root searches the default project and global locations for
		// every known harness rather than only project-local Codex skills.
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

func runTUI(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("skilltrace", flag.ContinueOnError)
	fs.SetOutput(stderr)
	catalogPath := fs.String("catalog", "", "catalog path")
	skillRoot := fs.String("skill-root", os.Getenv("SKILLTRACE_SKILL_ROOT"), "skill installation root (defaults to the project and global roots for every harness)")
	scanInput := fs.String("scan-input", os.Getenv("SKILLTRACE_SCAN_INPUT"), "trace JSONL path reachable with the scan key")
	harness := fs.String("harness", "codex", "harness for the scan input")
	policyFile := fs.String("policy", "", "policy YAML file reachable with the policy key")
	cohortLeft := fs.String("left", "", "left cohort definition for the comparison view")
	cohortRight := fs.String("right", "", "right cohort definition for the comparison view")
	skill := fs.String("skill", "", "skill to compare cohorts for")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	// The TUI needs a terminal on both ends. Failing here with a generic
	// message sent users looking for a bug that was really just a pipe.
	f, ok := stdout.(*os.File)
	if !ok || !isatty.IsTerminal(f.Fd()) {
		fmt.Fprintln(stderr, "skilltrace: standard output is not a terminal; use a headless command such as 'skilltrace discover --format json'")
		return 2
	}

	if *catalogPath == "" {
		*catalogPath = defaultCatalogPath()
	}
	c, err := catalog.Open(*catalogPath)
	if err != nil {
		fmt.Fprintf(stderr, "skilltrace: catalog is unavailable at %s: %v\n", *catalogPath, err)
		return 1
	}
	defer c.Close()

	color := isatty.IsTerminal(f.Fd()) && os.Getenv("TERM") != "dumb" && os.Getenv("NO_COLOR") == ""
	locale := strings.ToUpper(os.Getenv("LC_ALL") + os.Getenv("LC_CTYPE") + os.Getenv("LANG"))
	root, err := tui.Load(app.New(c), tui.Config{
		SkillRoot: *skillRoot, ScanInput: *scanInput, Harness: *harness,
		PolicyFile: *policyFile, CohortLeft: *cohortLeft, CohortRight: *cohortRight, Skill: *skill,
		Color: color, Unicode: strings.Contains(locale, "UTF-8") || strings.Contains(locale, "UTF8"),
	})
	if err != nil {
		fmt.Fprintf(stderr, "skilltrace: could not load discovery: %v\n", err)
		return 1
	}
	final, err := tea.NewProgram(root, tea.WithOutput(stdout)).Run()
	if err != nil {
		fmt.Fprintf(stderr, "skilltrace: terminal session failed: %v\n", err)
		return 1
	}
	if session, ok := final.(*tui.Root); ok && session.Interrupted {
		return 130
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
		result, err := a.Scan(ctx, app.ScanRequest{Input: *input, Harness: *harness}, func(p app.Progress) {
			_ = headless.RenderProgress(stderr, "source scan", p.Stage, p.Completed, p.Total)
		})
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
	ae := apperror.Public(err)
	code, message := ae.Code, ae.Message
	_ = headless.Render(w, format, headless.Failure(command, code, message))
	return headless.ExitCode(code)
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
