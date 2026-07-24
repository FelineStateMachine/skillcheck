package app

import (
	"context"
	"sync"
)

// Coordinator serializes commits independently from parallel parsing.
type Coordinator struct{ mu sync.Mutex }

func (c *Coordinator) Commit(ctx context.Context, fn func(context.Context) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn(ctx)
}
