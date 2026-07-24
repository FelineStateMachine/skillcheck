package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"skilltrace/internal/benchmark"
	"skilltrace/internal/headless"
)

func runBenchmark(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 || args[1] != "run" {
		fmt.Fprintln(stderr, "usage: skilltrace benchmark run --fixtures PATH")
		return 2
	}
	fs := flag.NewFlagSet("benchmark run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fixtures := fs.String("fixtures", "", "fixture directory")
	format := fs.String("format", "text", "output format")
	if err := fs.Parse(args[2:]); err != nil {
		return 2
	}
	if *fixtures == "" {
		fmt.Fprintln(stderr, "fixtures are required")
		return 2
	}
	report, err := benchmark.Run(ctx, *fixtures)
	if err != nil {
		return renderError(stdout, *format, "benchmark run", err)
	}
	if headless.Render(stdout, *format, headless.Success("benchmark run", report)) != nil {
		return 1
	}
	return 0
}
