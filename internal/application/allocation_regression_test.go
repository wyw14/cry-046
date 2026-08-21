package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
)

func TestAllocationRejectsNegativeAndSortsPlan(t *testing.T) {
	lines := []domain.AllocationLine{{BatchID: "b2", ProjectID: "p1", AmountCents: 600}, {BatchID: "b1", ProjectID: "p1", AmountCents: 400}, {BatchID: "bad", ProjectID: "p2", AmountCents: -50}}
	got, err := BuildAllocationResult(context.Background(), "p1", 1000, lines)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Plan.Balanced || got.Plan.RemainingCents != 0 {
		t.Fatalf("plan not balanced: %+v", got.Plan)
	}
	if got.Plan.Lines[0].BatchID != "b1" {
		t.Fatalf("lines not sorted: %+v", got.Plan.Lines)
	}
}
