package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
	"time"
)

func TestSummaryDigestIncludesRuleVersionAndFullEntryIdentity(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	entries := []domain.SettlementEntry{{ID: "e2", Amount: 100, Currency: "CNY", OccurredAt: now, SourceFingerprint: "fp2"}, {ID: "e1", Amount: 100, Currency: "CNY", OccurredAt: now, SourceFingerprint: "fp1"}}
	rv1 := domain.RuleVersion{ID: "rv1", Code: "R1"}
	rv2 := domain.RuleVersion{ID: "rv2", Code: "R2"}
	d1, err := BuildSummaryDigest(context.Background(), "c1", rv1, entries)
	if err != nil {
		t.Fatal(err)
	}
	d2, err := BuildSummaryDigest(context.Background(), "c1", rv2, entries)
	if err != nil {
		t.Fatal(err)
	}
	if d1.Digest == d2.Digest {
		t.Fatalf("rule version omitted from digest: %+v", d1)
	}
}
