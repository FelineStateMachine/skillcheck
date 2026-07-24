package policy

import "skilltrace/internal/app"

type Model struct {
	File       string
	Preview    app.PolicyPreview
	Validation app.PolicyValidation
	Err        string
	Applied    bool
}

func New(file string) Model { return Model{File: file} }
