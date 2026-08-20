package callback

import (
	"context"
	"testing"
)

func TestAdapterEmitAndSnapshot(t *testing.T) {
	a := New(3)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, err := a.Emit(ctx, "topic", map[string]any{"i": i})
		if err != nil {
			t.Fatalf("emit %d: %v", i, err)
		}
	}
	snap := a.Snapshot()
	if len(snap) != 3 {
		t.Errorf("expected 3 events (bounded), got %d", len(snap))
	}
	// Should be the last 3 (i=2,3,4).
	if snap[0].Payload["i"].(int) != 2 {
		t.Errorf("expected first event to be i=2, got %v", snap[0].Payload["i"])
	}
	if snap[2].Payload["i"].(int) != 4 {
		t.Errorf("expected last event to be i=4, got %v", snap[2].Payload["i"])
	}
}

func TestAdapterReset(t *testing.T) {
	a := New(10)
	ctx := context.Background()
	_, _ = a.Emit(ctx, "t", nil)
	if len(a.Snapshot()) != 1 {
		t.Fatal("expected 1 event before reset")
	}
	a.Reset()
	if len(a.Snapshot()) != 0 {
		t.Fatal("expected 0 events after reset")
	}
}

func TestAdapterCancelledContext(t *testing.T) {
	a := New(2)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := a.Emit(ctx, "topic", nil)
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}
