package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
)

var _ = domain.Adjustment{}

func TestApplyAdjustment_RejectsNegativeCumulativeDisbursement(t *testing.T) {
	app := NewSummaryApp(nil, nil, nil, nil, nil, nil, newAnnualRepoFake(), &auditRepoFake{}, newFakeClock())
	_, err := app.ApplyAdjustment(context.Background(), AdjustAnnualInput{TenantID: "t1", ProjectID: "p1", Year: 2026, DeltaCents: -1, Reason: "correction"})
	if err == nil {
		t.Fatal("negative cumulative disbursement must be rejected")
	}
}
