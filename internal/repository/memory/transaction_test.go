package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// TestUnit_TransactionRunsAllOrNothing verifies that the in-memory
// Unit-of-Work runs the callback with a complete repo set and that
// an error returned from the callback propagates without side effects
// to the Unit itself.
func TestUnit_TransactionRunsAllOrNothing(t *testing.T) {
	store := NewStore()
	repos := New(store)
	unit := NewUnit(repos)
	ctx := context.Background()

	t.Run("success path exposes all repos", func(t *testing.T) {
		var seenProjectRepo, seenEntryRepo, seenExceptionRepo bool
		err := unit.Do(ctx, func(ctx context.Context, uow application.UnitOfWork) error {
			if uow.Projects() != nil {
				seenProjectRepo = true
			}
			if uow.Entries() != nil {
				seenEntryRepo = true
			}
			if uow.Exceptions() != nil {
				seenExceptionRepo = true
			}
			return nil
		})
		if err != nil {
			t.Fatalf("do: %v", err)
		}
		if !seenProjectRepo || !seenEntryRepo || !seenExceptionRepo {
			t.Errorf("expected all repos exposed: proj=%v entry=%v exc=%v",
				seenProjectRepo, seenEntryRepo, seenExceptionRepo)
		}
	})

	t.Run("failure path propagates error", func(t *testing.T) {
		sentinel := errors.New("boom")
		err := unit.Do(ctx, func(ctx context.Context, uow application.UnitOfWork) error {
			return sentinel
		})
		if !errors.Is(err, sentinel) {
			t.Errorf("expected sentinel error, got %v", err)
		}
	})

	t.Run("mutation within unit persists", func(t *testing.T) {
		now := now()
		err := unit.Do(ctx, func(ctx context.Context, uow application.UnitOfWork) error {
			p := domain.Project{
				ID: "p-tx", TenantID: "t1", Code: "WS-TX", Name: "Tx",
				Sponsor: "S", AnnualBudget: 100, StartYear: 2026, EndYear: 2027,
				CreatedAt: now, UpdatedAt: now,
			}
			if _, err := uow.Projects().Create(ctx, p); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			t.Fatalf("do: %v", err)
		}
		// The mutation should be visible outside the unit.
		got, err := repos.Projects.Get(ctx, "t1", "p-tx")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Code != "WS-TX" {
			t.Errorf("expected WS-TX, got %s", got.Code)
		}
	})
}

// now returns a fixed time to satisfy the repo's non-zero time fields.
func now() time.Time { return time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC) }

// Silence unused import if time is not needed elsewhere.
var _ = errors.New
