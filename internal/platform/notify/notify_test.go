package notify

import (
	"context"
	"testing"
)

func TestAdapterSendAndSnapshot(t *testing.T) {
	a := New(3)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, err := a.Send(ctx, Message{
			Recipient: "u1", Channel: "system", Subject: "test", Body: "body",
		})
		if err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}
	snap := a.Snapshot()
	if len(snap) != 3 {
		t.Errorf("expected 3 messages (bounded), got %d", len(snap))
	}
	for _, m := range snap {
		if m.DeliveredAt.IsZero() {
			t.Error("expected DeliveredAt to be set")
		}
	}
}

func TestAdapterReset(t *testing.T) {
	a := New(10)
	ctx := context.Background()
	_, _ = a.Send(ctx, Message{Recipient: "u1"})
	if len(a.Snapshot()) != 1 {
		t.Fatal("expected 1 message before reset")
	}
	a.Reset()
	if len(a.Snapshot()) != 0 {
		t.Fatal("expected 0 messages after reset")
	}
}

func TestAdapterCancelledContext(t *testing.T) {
	a := New(2)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := a.Send(ctx, Message{Recipient: "u1"})
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}
