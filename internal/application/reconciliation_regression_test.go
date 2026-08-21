package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
	"time"
)

func TestReconcileEntriesDetectsSourceVariants(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	existing := []domain.SettlementEntry{{ID: "old", TenantID: "t1", SourceID: "bank-1", SourceFingerprint: "oldfp", Amount: 100, Currency: "CNY", OccurredAt: now}}
	incoming := []domain.SettlementEntry{{ID: "same", TenantID: "t1", SourceID: "bank-1", SourceFingerprint: "fp1", Amount: 100, Currency: "CNY", OccurredAt: now}, {ID: "variant", TenantID: "t1", SourceID: "bank-1", SourceFingerprint: "fp2", Amount: 120, Currency: "CNY", OccurredAt: now}}
	got, err := ReconcileEntries(context.Background(), "t1", existing, incoming)
	if err != nil {
		t.Fatal(err)
	}
	if got.Report.Conflicts != 1 || len(got.Rejected) != 1 {
		t.Fatalf("source variants not detected: %+v", got)
	}
}
