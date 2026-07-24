package main

import (
	"flag"
	"io"
	"os"

	"skilltrace/internal/apperror"
	"skilltrace/internal/headless"
	"skilltrace/internal/policy"
)

func runCorrection(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		return 2
	}
	fs := flag.NewFlagSet("correction "+args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)
	file := fs.String("file", "", "policy YAML file")
	id := fs.String("id", "", "correction id")
	format := fs.String("format", "text", "output format")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if *file == "" || *id == "" {
		return renderError(stdout, *format, "correction "+args[0], apperror.Wrap("invalid_arguments", "file and id are required", nil))
	}
	doc, diagnostics, err := (policy.Store{}).Read(*file)
	if err != nil || len(diagnostics) > 0 {
		return renderError(stdout, *format, "correction "+args[0], apperror.Wrap("policy_error", "policy is invalid", err))
	}
	value := args[0] == "enable"
	if args[0] != "enable" && args[0] != "suppress" {
		return 2
	}
	updated, err := policy.SetEnabled(doc.Original, "correction", *id, value)
	if err != nil {
		return renderError(stdout, *format, "correction "+args[0], apperror.Wrap("policy_conflict", "policy edit could not be applied", err))
	}
	if _, err = (policy.Store{}).Replace(*file, doc.Fingerprint, updated); err != nil {
		return renderError(stdout, *format, "correction "+args[0], apperror.Wrap("policy_conflict", "policy changed since it was read", err))
	}
	_ = os.Chmod(*file, 0o600)
	if err := headless.Render(stdout, *format, headless.Success("correction "+args[0], headless.PolicyMutationResult{Updated: true, File: *file})); err != nil {
		return 1
	}
	return 0
}
