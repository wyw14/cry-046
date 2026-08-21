package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
	"time"
)

func TestSLAQueueExcludesResolvedAndZeroDeadline(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	xs := []domain.Exception{{ID: "open", TenantID: "t1", Status: domain.ExceptionStatusProcessing, DeadlineAt: now.Add(-time.Hour)}, {ID: "resolved", TenantID: "t1", Status: domain.ExceptionStatusResolved, DeadlineAt: now.Add(-time.Hour)}, {ID: "no-deadline", TenantID: "t1", Status: domain.ExceptionStatusPending}}
	got, err := BuildSLAQueue(context.Background(), "t1", xs, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.Report.OverdueCount != 1 || got.Report.OpenCount != 2 || len(got.Escalate) != 1 {
		t.Fatalf("SLA lifecycle wrong: %+v", got)
	}
}
