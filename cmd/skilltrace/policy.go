package main

import (
	"flag"
	"fmt"
	"io"

	"skilltrace/internal/app"
	"skilltrace/internal/apperror"
	"skilltrace/internal/catalog"
	"skilltrace/internal/headless"
)

func runPolicy(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: skilltrace policy <validate|preview|apply>")
		return 2
	}
	fs := flag.NewFlagSet("policy "+args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)
	file := fs.String("file", "", "policy YAML file")
	format := fs.String("format", "text", "output format")
	catalogPath := fs.String("catalog", "", "catalog path")
	token := fs.String("token", "", "revision-bound preview token")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if *file == "" {
		return renderError(stdout, *format, "policy "+args[0], apperror.Wrap("invalid_arguments", "file is required", nil))
	}
	if *catalogPath == "" {
		*catalogPath = defaultCatalogPath()
	}
	c, err := catalog.Open(*catalogPath)
	if err != nil {
		return renderError(stdout, *format, "policy "+args[0], apperror.Wrap("catalog_unavailable", "catalog is unavailable", err))
	}
	defer c.Close()
	a := app.New(c)
	var result any
	switch args[0] {
	case "validate":
		result, err = a.ValidatePolicy(*file)
	case "preview":
		result, err = a.PreviewPolicy(*file)
	case "apply":
		result, err = a.ApplyPolicy(*file, *token)
	default:
		fmt.Fprintln(stderr, "unknown policy command")
		return 2
	}
	if err != nil {
		return renderError(stdout, *format, "policy "+args[0], apperror.Wrap("policy_error", "policy operation failed", err))
	}
	if err := headless.Render(stdout, *format, headless.Success("policy "+args[0], result)); err != nil {
		return 1
	}
	return 0
}
