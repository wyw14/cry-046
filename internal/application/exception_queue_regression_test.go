package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
	"time"
)

func TestExceptionQueueOrdersSeverityThenKnownDeadline(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	xs := []domain.Exception{{ID: "low-early", TenantID: "t1", Severity: domain.SeverityLow, DeadlineAt: now.Add(time.Hour), Status: domain.ExceptionStatusPending}, {ID: "high-none", TenantID: "t1", Severity: domain.SeverityHigh, Status: domain.ExceptionStatusPending}, {ID: "high-late", TenantID: "t1", Severity: domain.SeverityHigh, DeadlineAt: now.Add(3 * time.Hour), Status: domain.ExceptionStatusPending}, {ID: "high-early", TenantID: "t1", Severity: domain.SeverityHigh, DeadlineAt: now.Add(2 * time.Hour), Status: domain.ExceptionStatusPending}}
	got, err := BuildExceptionQueue(context.Background(), "t1", xs, now)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"high-early", "high-late", "high-none", "low-early"}
	for i, id := range want {
		if got.Items[i].ID != id {
			t.Fatalf("order[%d]=%s want %s: %+v", i, got.Items[i].ID, id, got.Items)
		}
	}
}
