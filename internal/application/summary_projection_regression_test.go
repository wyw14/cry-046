package application

import (
	"context"
	"testing"
	"time"

	"github.com/welfare/settlement-resolver/internal/domain"
)

func TestSummaryProjectionPreservesPrecisionAndLatestVersion(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	history := []domain.Summary{
		{ID: "old", CycleID: "c1", RuleVersionID: "r1", ComputedAt: now, Version: 2, TotalAmountCents: 3, ApprovedAmountCents: 1, PendingAmountCents: 2},
		{ID: "latest", CycleID: "c1", RuleVersionID: "r1", ComputedAt: now.Add(time.Minute), Version: 3, TotalAmountCents: 8, ApprovedAmountCents: 5, PendingAmountCents: 3},
	}
	got, err := BuildSummaryProjection(context.Background(), "c1", history)
	if err != nil {
		t.Fatal(err)
	}
	if got.LatestVersion != 3 || got.ApprovedBPS != 6250 || got.PendingBPS != 3750 {
		t.Fatalf("latest projection mismatch: %+v", got)
	}
	if len(got.Points) != 2 || got.Points[0].SummaryID != "latest" || got.Points[1].ApprovedBPS != 3333 {
		t.Fatalf("history precision/order mismatch: %+v", got.Points)
	}
}
