package app

import (
	"context"
	"testing"
)

func TestSchedulerDeterministicCompletionOrder(t *testing.T) {
	work := []Work{{Key: "b", Priority: 0, Bytes: 1, Run: func(context.Context) (any, error) { return "b", nil }}, {Key: "a", Priority: 1, Bytes: 1, Run: func(context.Context) (any, error) { return "a", nil }}}
	got := RunScheduled(context.Background(), 2, 10, work)
	if got[0].Key != "a" || got[1].Key != "b" {
		t.Fatalf("unexpected order: %#v", got)
	}
}

func TestSchedulerRecoversPanic(t *testing.T) {
	got := RunScheduled(context.Background(), 1, 10, []Work{{Key: "panic", Bytes: 1, Run: func(context.Context) (any, error) { panic("secret") }}})
	if got[0].Err == nil || got[0].Err.Error() != "worker panic recovered" {
		t.Fatalf("unexpected error: %v", got[0].Err)
	}
}
