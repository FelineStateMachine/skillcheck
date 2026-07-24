package policy

import (
	"crypto/sha256"
	"fmt"

	"github.com/goccy/go-yaml"
)

type Document struct {
	V1          DocumentV1 `json:"policy"`
	Original    []byte     `json:"-"`
	Fingerprint string     `json:"fingerprint"`
}

func Parse(data []byte) (*Document, []Diagnostic) {
	var dto DocumentV1
	if err := yaml.UnmarshalWithOptions(data, &dto, yaml.Strict()); err != nil {
		return nil, []Diagnostic{{Code: "invalid_yaml", Message: safeYAMLError(err)}}
	}
	if diagnostics := Validate(dto); len(diagnostics) != 0 {
		return nil, diagnostics
	}
	sum := sha256.Sum256(data)
	return &Document{V1: dto, Original: append([]byte(nil), data...), Fingerprint: fmt.Sprintf("%x", sum[:])}, nil
}

func safeYAMLError(err error) string {
	// Parser diagnostics can contain scalar values and paths. Keep ordinary output
	// on the stable contract rather than forwarding potentially sensitive input.
	if err == nil {
		return "invalid YAML policy"
	}
	return "invalid YAML policy syntax or schema"
}
