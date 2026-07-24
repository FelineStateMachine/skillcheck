package catalog_test

import (
	"context"
	"testing"

	"skilltrace/internal/adapters"
	"skilltrace/internal/catalog"
	"skilltrace/internal/trace"
)

func TestReplaceSourceRollsBackAtomically(t *testing.T) {
	c, err := catalog.Open(t.TempDir() + "/catalog.db")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	e, _ := trace.NewEvent("session", 1, 1, trace.SessionPayload{Model: "safe"})
	base := adapters.Result{Harness: "codex", Events: []trace.Event{e}, Capabilities: trace.CodexCapabilities()}
	if _, err = c.ReplaceSource(context.Background(), "source", base); err != nil {
		t.Fatal(err)
	}
	bad := base
	bad.Events = []trace.Event{e, e}
	if _, err = c.ReplaceSource(context.Background(), "source", bad); err == nil {
		t.Fatal("expected duplicate sequence failure")
	}
	health, err := c.Health()
	if err != nil {
		t.Fatal(err)
	}
	if len(health) != 1 || health[0].Revision != 1 || health[0].EventCount != 1 {
		t.Fatalf("prior snapshot changed: %+v", health)
	}
}
