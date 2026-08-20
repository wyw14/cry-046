package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// TestMemoryRepo_ConcurrentExceptionAssign verifies that the in-memory
// store serialises concurrent updates to the same exception and that
// optimistic concurrency is enforced: only the goroutine whose
// read-then-update lands first against the current version succeeds;
// every other goroutine observing the same baseline must abort.
//
// To make the race deterministic the goroutines all read the seed
// version BEFORE any of them write, then attempt the update. Because
// the in-memory store uses a single mutex, the writes are serialised;
// the first one to acquire the lock sees Version=1 and accepts the
// incoming Version=2. Subsequent writers see Version=2 and must abort
// because their incoming Version is still 2 (expected 3).
func TestMemoryRepo_ConcurrentExceptionAssign(t *testing.T) {
	store := NewStore()
	repo := &ExceptionRepo{store: store}
	ctx := context.Background()
	now := time.Now()

	seed := domain.Exception{
		ID: "ex-c", TenantID: "t1", CycleID: "c1", EntryID: "e1", RuleVersionID: "rv1",
		RuleCode: "X", Severity: domain.SeverityHigh, Title: "x", Status: domain.ExceptionStatusPending,
		Version: 1, Snapshot: domain.ExceptionSnapshot{SnapshotAt: now, EntryAmountCents: 0, EntryCurrency: "CNY"},
	}
	if _, err := repo.Create(ctx, seed); err != nil {
		t.Fatalf("create: %v", err)
	}

	// All goroutines read the seed version up-front so they all
	// observe Version=1 and attempt to write Version=2.
	type payload struct {
		ex  domain.Exception
		err error
	}
	results := make([]payload, 50)
	readBarrier := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ex, err := repo.Get(ctx, "t1", "ex-c")
			results[i] = payload{ex: ex, err: err}
			<-readBarrier // wait until all goroutines have read
			ex.AssigneeID = fmt.Sprintf("assignee-%d", i)
			ex.Version = ex.Version + 1
			_, err = repo.Update(ctx, ex)
			results[i].err = err
		}(i)
	}
	// Give goroutines a moment to complete their reads, then release.
	time.Sleep(20 * time.Millisecond)
	close(readBarrier)
	wg.Wait()

	var success, failed int32
	for _, r := range results {
		if r.err == nil {
			success++
		} else {
			failed++
		}
	}
	if success+failed != 50 {
		t.Errorf("expected 50 total, got success=%d failed=%d", success, failed)
	}
	if success != 1 {
		t.Errorf("expected exactly 1 success, got %d (failed=%d)", success, failed)
	}
	if failed != 49 {
		t.Errorf("expected 49 failed, got %d", failed)
	}
}

func TestMemoryRepo_FilterExceptions(t *testing.T) {
	store := NewStore()
	repo := &ExceptionRepo{store: store}
	ctx := context.Background()
	now := time.Now()

	exs := []domain.Exception{
		{ID: "e1", TenantID: "t1", CycleID: "c1", EntryID: "x1", RuleVersionID: "rv", RuleCode: "R", Severity: domain.SeverityCritical, Title: "x", Status: domain.ExceptionStatusPending, Version: 1, Snapshot: domain.ExceptionSnapshot{SnapshotAt: now}, DeadlineAt: now.Add(-time.Hour)},
		{ID: "e2", TenantID: "t1", CycleID: "c1", EntryID: "x2", RuleVersionID: "rv", RuleCode: "R", Severity: domain.SeverityLow, Title: "x", Status: domain.ExceptionStatusProcessing, Version: 1, Snapshot: domain.ExceptionSnapshot{SnapshotAt: now}},
		{ID: "e3", TenantID: "t1", CycleID: "c1", EntryID: "x3", RuleVersionID: "rv", RuleCode: "R", Severity: domain.SeverityHigh, Title: "x", Status: domain.ExceptionStatusResolved, Version: 1, Snapshot: domain.ExceptionSnapshot{SnapshotAt: now}},
	}
	for _, e := range exs {
		if _, err := repo.Create(ctx, e); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	// Filter by status.
	list, _, err := repo.List(ctx, application.ExceptionListQuery{
		ListQuery: application.ListQuery{TenantID: "t1", PageSize: 100},
		Status:    string(domain.ExceptionStatusPending),
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].ID != "e1" {
		t.Errorf("expected e1, got %+v", list)
	}

	// Overdue-only.
	list, _, err = repo.List(ctx, application.ExceptionListQuery{
		ListQuery:   application.ListQuery{TenantID: "t1", PageSize: 100},
		OverdueOnly: true,
		AsOf:        now,
	})
	if err != nil {
		t.Fatalf("list overdue: %v", err)
	}
	if len(list) != 1 || list[0].ID != "e1" {
		t.Errorf("expected only e1 overdue, got %+v", list)
	}

	// Severity filter.
	list, _, err = repo.List(ctx, application.ExceptionListQuery{
		ListQuery: application.ListQuery{TenantID: "t1", PageSize: 100},
		Severity:  string(domain.SeverityLow),
	})
	if err != nil {
		t.Fatalf("list severity: %v", err)
	}
	if len(list) != 1 || list[0].ID != "e2" {
		t.Errorf("expected e2, got %+v", list)
	}
}

func TestMemoryRepo_UserUniqueName(t *testing.T) {
	store := NewStore()
	repo := &UserRepo{store: store}
	ctx := context.Background()
	now := time.Now()

	u1 := domain.User{ID: "u1", TenantID: "t1", Username: "admin", Role: domain.RoleAdmin, CreatedAt: now, UpdatedAt: now}
	if _, err := repo.Create(ctx, u1); err != nil {
		t.Fatalf("create: %v", err)
	}
	// Duplicate username → error.
	u2 := domain.User{ID: "u2", TenantID: "t1", Username: "admin", Role: domain.RoleAdmin, CreatedAt: now, UpdatedAt: now}
	_, err := repo.Create(ctx, u2)
	if err == nil || !domain.IsAlreadyExists(err) {
		t.Fatalf("expected AlreadyExists, got %v", err)
	}
	// Different tenant, same name → allowed.
	u3 := domain.User{ID: "u3", TenantID: "t2", Username: "admin", Role: domain.RoleAdmin, CreatedAt: now, UpdatedAt: now}
	if _, err := repo.Create(ctx, u3); err != nil {
		t.Fatalf("create in different tenant: %v", err)
	}
}

func TestMemoryRepo_RuleVersionOptimisticLock(t *testing.T) {
	store := NewStore()
	repo := &RuleVersionRepo{store: store}
	ctx := context.Background()
	now := time.Now()

	rv := domain.RuleVersion{
		ID: "rv1", TenantID: "t1", Code: "RV-1", ProjectID: "p1",
		Rules:   []domain.RuleDefinition{{Code: "X", Severity: domain.SeverityLow, Expression: "amount == 0"}},
		Status:  domain.RuleVersionStatusDraft,
		Version: 1, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := repo.Create(ctx, rv); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Stale version.
	rv.Description = "updated"
	rv.Version = 1 // unchanged
	if _, err := repo.Update(ctx, rv); err == nil {
		t.Fatal("expected aborted for stale version")
	}

	// Correct version.
	rv.Version = 2
	if _, err := repo.Update(ctx, rv); err != nil {
		t.Fatalf("update: %v", err)
	}
}

func TestMemoryRepo_PageSlice(t *testing.T) {
	in := make([]int, 100)
	for i := range in {
		in[i] = i
	}
	got := pageSlice(in, 2, 10)
	if len(got) != 10 || got[0] != 10 {
		t.Errorf("page 2 size 10: got %v", got)
	}
	got = pageSlice(in, 1, 1000)
	if len(got) != 100 {
		t.Errorf("page 1 size 1000: got %d", len(got))
	}
	got = pageSlice(in, 100, 10)
	if len(got) != 0 {
		t.Errorf("page 100: expected empty, got %d", len(got))
	}
}
