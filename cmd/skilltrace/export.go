package main

import (
	"context"
	"flag"
	"fmt"
	"io"

	"skilltrace/internal/app"
	"skilltrace/internal/catalog"
	"skilltrace/internal/headless"
)

func runExport(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "html" {
		fmt.Fprintln(stderr, "usage: skilltrace export html [options]")
		return 2
	}
	fs := flag.NewFlagSet("export html", flag.ContinueOnError)
	fs.SetOutput(stderr)
	format := fs.String("format", "text", "output format")
	catalogPath := fs.String("catalog", "", "catalog path")
	skill := fs.String("skill", "", "skill name")
	left := fs.String("left", "", "left cohort file")
	right := fs.String("right", "", "right cohort file")
	output := fs.String("output", "", "HTML destination")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if *catalogPath == "" {
		*catalogPath = defaultCatalogPath()
	}
	if *skill == "" || *left == "" || *right == "" || *output == "" {
		return renderError(stdout, *format, "export html", fmt.Errorf("skill, left, right, and output are required"))
	}
	c, err := catalog.Open(*catalogPath)
	if err != nil {
		return renderError(stdout, *format, "export html", err)
	}
	defer c.Close()
	result, err := app.New(c).ExportHTML(ctx, app.ExportRequest{Skill: *skill, LeftPath: *left, RightPath: *right, Output: *output})
	if err != nil {
		return renderError(stdout, *format, "export html", err)
	}
	if *format == "text" {
		_, err = fmt.Fprintln(stdout, headless.ExportText(result))
	} else {
		err = headless.Render(stdout, *format, headless.Success("export html", result))
	}
	if err != nil {
		return 1
	}
	return 0
}
