package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/welfare/settlement-resolver/internal/domain"
)

func makeException() domain.Exception {
	return domain.Exception{
		ID:            "ex-1",
		TenantID:      "t1",
		CycleID:       "c1",
		EntryID:       "e1",
		RuleVersionID: "rv1",
		RuleCode:      "AMOUNT_ZERO",
		Severity:      domain.SeverityHigh,
		Title:         "AMOUNT_ZERO: e1",
		Status:        domain.ExceptionStatusPending,
		Version:       1,
		Snapshot: domain.ExceptionSnapshot{
			SnapshotAt:       time.Now(),
			EntryAmountCents: 0,
			EntryCurrency:    "CNY",
		},
	}
}

func TestExceptionsApp_AssignClaimWorkflow(t *testing.T) {
	clk := newFakeClock()
	excs := &exceptionRepoFake{items: []domain.Exception{makeException()}}
	audit := &auditRepoFake{}
	app := NewExceptionsApp(excs, audit, clk)

	t.Run("assign transitions to processing", func(t *testing.T) {
		out, err := app.Assign(context.Background(), AssignInput{
			TenantID: "t1", ExceptionID: "ex-1", AssigneeID: "assignee-1", AuthorID: "admin", Note: "请处理",
		})
		if err != nil {
			t.Fatalf("assign: %v", err)
		}
		if out.Status != domain.ExceptionStatusProcessing {
			t.Errorf("expected processing, got %s", out.Status)
		}
		if out.AssigneeID != "assignee-1" {
			t.Errorf("expected assignee assignee-1, got %s", out.AssigneeID)
		}
		if out.Version != 2 {
			t.Errorf("expected version 2, got %d", out.Version)
		}
		if len(out.Notes) != 1 || out.Notes[0].Kind != domain.NoteKindAssignment {
			t.Errorf("expected assignment note, got %+v", out.Notes)
		}
	})

	t.Run("claim by different assignee fails after assign", func(t *testing.T) {
		excs2 := &exceptionRepoFake{items: []domain.Exception{makeException()}}
		app2 := NewExceptionsApp(excs2, audit, clk)
		_, err := app2.Assign(context.Background(), AssignInput{
			TenantID: "t1", ExceptionID: "ex-1", AssigneeID: "assignee-A",
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = app2.Claim(context.Background(), ClaimInput{
			TenantID: "t1", ExceptionID: "ex-1", AssigneeID: "assignee-B",
		})
		if err == nil || !strings.Contains(err.Error(), "already claimed") {
			t.Fatalf("expected already-claimed error, got %v", err)
		}
	})
}

func TestExceptionsApp_ResolveSeparationOfDuties(t *testing.T) {
	clk := newFakeClock()

	t.Run("reviewer == assignee rejected", func(t *testing.T) {
		ex := makeException()
		ex.Status = domain.ExceptionStatusReview
		ex.AssigneeID = "assignee-1"
		excs := &exceptionRepoFake{items: []domain.Exception{ex}}
		app := NewExceptionsApp(excs, &auditRepoFake{}, clk)
		_, err := app.Resolve(context.Background(), ResolveInput{
			TenantID: "t1", ExceptionID: "ex-1", ReviewerID: "assignee-1",
		})
		if err == nil || !strings.Contains(err.Error(), "reviewer must differ") {
			t.Fatalf("expected separation-of-duties error, got %v", err)
		}
	})

	t.Run("reviewer != assignee resolves", func(t *testing.T) {
		ex := makeException()
		ex.Status = domain.ExceptionStatusReview
		ex.AssigneeID = "assignee-1"
		excs := &exceptionRepoFake{items: []domain.Exception{ex}}
		app := NewExceptionsApp(excs, &auditRepoFake{}, clk)
		out, err := app.Resolve(context.Background(), ResolveInput{
			TenantID: "t1", ExceptionID: "ex-1", ReviewerID: "reviewer-1",
		})
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if out.Status != domain.ExceptionStatusResolved {
			t.Errorf("expected resolved, got %s", out.Status)
		}
		if out.ResolvedAt.IsZero() {
			t.Error("ResolvedAt must be set")
		}
	})
}

func TestExceptionsApp_EscalateAndRework(t *testing.T) {
	clk := newFakeClock()

	t.Run("escalate pending", func(t *testing.T) {
		excs := &exceptionRepoFake{items: []domain.Exception{makeException()}}
		app := NewExceptionsApp(excs, &auditRepoFake{}, clk)
		out, err := app.Escalate(context.Background(), EscalateInput{
			TenantID: "t1", ExceptionID: "ex-1", AuthorID: "operator-1", Reason: "无法定位收据",
		})
		if err != nil {
			t.Fatalf("escalate: %v", err)
		}
		if out.Status != domain.ExceptionStatusEscalated {
			t.Errorf("expected escalated, got %s", out.Status)
		}
	})

	t.Run("escalate requires reason", func(t *testing.T) {
		excs := &exceptionRepoFake{items: []domain.Exception{makeException()}}
		app := NewExceptionsApp(excs, &auditRepoFake{}, clk)
		_, err := app.Escalate(context.Background(), EscalateInput{
			TenantID: "t1", ExceptionID: "ex-1", AuthorID: "operator-1",
		})
		if err == nil || !strings.Contains(err.Error(), "reason must not be empty") {
			t.Fatalf("expected reason error, got %v", err)
		}
	})

	t.Run("rework resolved requires note", func(t *testing.T) {
		ex := makeException()
		ex.Status = domain.ExceptionStatusResolved
		ex.AssigneeID = "a"
		excs := &exceptionRepoFake{items: []domain.Exception{ex}}
		app := NewExceptionsApp(excs, &auditRepoFake{}, clk)
		_, err := app.Rework(context.Background(), ReworkInput{
			TenantID: "t1", ExceptionID: "ex-1", AuthorID: "reviewer-1",
		})
		if err == nil || !strings.Contains(err.Error(), "rework note must not be empty") {
			t.Fatalf("expected rework note error, got %v", err)
		}
	})
}

func TestExceptionsApp_OptimisticConcurrency(t *testing.T) {
	clk := newFakeClock()
	excs := &exceptionRepoFake{items: []domain.Exception{makeException()}, staleVer: true}
	audit := &auditRepoFake{}
	app := NewExceptionsApp(excs, audit, clk)

	// Force a stale update by sending an unchanged version.
	// Stale-version logic in domain.AssignException bumps version to 2;
	// the repo will see existing.Version=1 and demand incoming.Version==2.
	// Pass it 1 to simulate a lost update.
	ex, _ := app.Get(context.Background(), "t1", "ex-1")
	// Intentionally do not increment Version — simulate a stale write
	// that lost the increment race.
	_, err := excs.Update(context.Background(), ex)
	if err == nil || !domain.IsAborted(err) {
		t.Fatalf("expected aborted, got %v", err)
	}
}

// TestExceptionsApp_ConcurrentAssign verifies that two concurrent
// assignments to the same exception cannot both succeed: one must
// observe the updated version and abort. The test reads the exception
// once (capturing Version=N), then tries to update it twice. The
// first update succeeds (Version=N+1); the second, still using
// Version=N, must be aborted by the repository.
func TestExceptionsApp_ConcurrentAssign(t *testing.T) {
	clk := newFakeClock()
	excs := &exceptionRepoFake{
		items:    []domain.Exception{makeException()},
		staleVer: true,
	}
	audit := &auditRepoFake{}
	app := NewExceptionsApp(excs, audit, clk)

	// Read once.
	ex, err := app.Get(context.Background(), "t1", "ex-1")
	if err != nil {
		t.Fatal(err)
	}
	staleVersion := ex.Version

	// First assign succeeds.
	_, err = app.Assign(context.Background(), AssignInput{
		TenantID: "t1", ExceptionID: "ex-1",
		AssigneeID: "assignee-A", AuthorID: "admin",
	})
	if err != nil {
		t.Fatalf("first assign should succeed, got %v", err)
	}

	// Now try to apply a stale update directly via the repository.
	// The stale version (staleVersion) must be rejected because the
	// repository expects the next version to be staleVersion+1, but the
	// current stored version is staleVersion+1.
	staleEx := ex
	staleEx.Version = staleVersion + 1 // what AssignException would produce from the stale read
	_, err = excs.Update(context.Background(), staleEx)
	if err == nil || !domain.IsAborted(err) {
		t.Fatalf("expected aborted for stale update, got %v", err)
	}
}
