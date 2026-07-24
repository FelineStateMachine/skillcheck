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

func runCompare(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	fs.SetOutput(stderr)
	format := fs.String("format", "text", "output format: text or json")
	catalogPath := fs.String("catalog", "", "catalog path")
	skill := fs.String("skill", "", "skill name")
	left := fs.String("left", "", "left cohort definition")
	right := fs.String("right", "", "right cohort definition")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *skill == "" || *left == "" || *right == "" {
		return renderError(stdout, *format, "compare", apperror.Wrap("invalid_arguments", "skill, left, and right are required", nil))
	}
	if *catalogPath == "" {
		*catalogPath = defaultCatalogPath()
	}
	c, err := catalog.Open(*catalogPath)
	if err != nil {
		return renderError(stdout, *format, "compare", apperror.Wrap("catalog_unavailable", "catalog is unavailable", err))
	}
	defer c.Close()
	result, err := app.New(c).Compare(ctx, app.CompareRequest{Skill: *skill, LeftPath: *left, RightPath: *right})
	if err != nil {
		return renderError(stdout, *format, "compare", err)
	}
	if *format == "text" {
		_, err = io.WriteString(stdout, headless.CompareText(result)+"\n")
	} else {
		err = headless.Render(stdout, *format, headless.Success("compare", result))
	}
	if err != nil {
		return 1
	}
	return 0
}
