package requestid

import (
	"context"
	"testing"
)

func TestNewIsUnique(t *testing.T) {
	a := New()
	b := New()
	if a == "" {
		t.Error("expected non-empty id")
	}
	if a == b {
		t.Error("expected unique ids")
	}
}

func TestWithFromRoundTrip(t *testing.T) {
	id := "req-123"
	ctx := With(context.Background(), id)
	if got := From(ctx); got != id {
		t.Errorf("expected %s, got %s", id, got)
	}
}

func TestFromEmpty(t *testing.T) {
	if got := From(context.Background()); got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}
