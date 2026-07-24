package adapters

import (
	"context"
	"fmt"
	"io"
)

type Registry struct{ adapters map[string]Adapter }

func NewRegistry() *Registry                                 { return &Registry{adapters: map[string]Adapter{}} }
func (r *Registry) Register(harness string, adapter Adapter) { r.adapters[harness] = adapter }
func (r *Registry) Parse(ctx context.Context, harness string, input io.Reader) (Result, error) {
	a, ok := r.adapters[harness]
	if !ok {
		return Result{}, fmt.Errorf("unsupported harness %q", harness)
	}
	return a.Parse(ctx, input)
}
func (r *Registry) Supports(harness string) bool { _, ok := r.adapters[harness]; return ok }
