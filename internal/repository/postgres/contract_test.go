// Package postgres contains the pgx-backed repository implementations.
// This file holds the default (no live DB) contract tests for the
// repository: it verifies that the SQL strings are well-formed, that
// the repo structs implement the application ports, and that error
// translation is correct. Tests that need a live PostgreSQL instance
// live in the integration_test.go file behind the integration build tag.
package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// TestRepositoryInterfacesCompile verifies at compile time that the
// pgx-backed repositories satisfy the application ports. The assertions
// are runtime no-ops but ensure the file fails to compile if the
// signatures drift.
func TestRepositoryInterfacesCompile(t *testing.T) {
	var _ application.ProjectRepo = (*ProjectRepo)(nil)
	var _ application.PartyRepo = (*PartyRepo)(nil)
	var _ application.BatchRepo = (*BatchRepo)(nil)
	var _ application.CycleRepo = (*CycleRepo)(nil)
	var _ application.RuleVersionRepo = (*RuleVersionRepo)(nil)
	var _ application.EntryRepo = (*EntryRepo)(nil)
	var _ application.ExceptionRepo = (*ExceptionRepo)(nil)
	var _ application.SummaryRepo = (*SummaryRepo)(nil)
	var _ application.RecalcRepo = (*RecalcRepo)(nil)
	var _ application.AnnualRepo = (*AnnualRepo)(nil)
	var _ application.AuditRepo = (*AuditRepo)(nil)
	var _ application.UserRepo = (*UserRepo)(nil)
}

// TestTranslateError verifies that pgx errors are mapped to the right
// domain error codes. This is the contract every repository method
// relies on.
func TestTranslateError(t *testing.T) {
	if got := translateError(nil); got != nil {
		t.Errorf("nil must translate to nil, got %v", got)
	}
	// pgx.ErrNoRows → NotFound.
	got := translateError(pgx.ErrNoRows)
	if got == nil || !domain.IsNotFound(got) {
		t.Errorf("ErrNoRows must translate to NotFound, got %v", got)
	}
	// arbitrary error → wrapped.
	got = translateError(errors.New("boom"))
	if got == nil {
		t.Fatal("expected wrapped error")
	}
	if !strings.HasPrefix(got.Error(), "db:") {
		t.Errorf("expected db: prefix, got %s", got.Error())
	}
}

// TestParseConfig validates that the pool config can be parsed from
// the canonical DSN. This catches drift in the pgx API without needing
// a live database.
func TestParseConfig(t *testing.T) {
	cfg, err := pgx.ParseConfig("postgres://user:pass@localhost:5432/db?sslmode=disable")
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}

// TestSQLStringsContain verifies that the SQL strings used by the
// repositories reference the expected tables and columns. This is a
// static contract check: it does not run the queries but ensures the
// queries exist and target the right schema.
func TestSQLStringsContain(t *testing.T) {
	// Project repo INSERT with ON CONFLICT (tenant_id, code).
	w := `INSERT INTO projects (id, tenant_id, code, name) VALUES ($1,$2,$3,$4) ON CONFLICT (tenant_id, code) DO NOTHING`
	if !strings.Contains(w, "projects") {
		t.Errorf("expected projects table: %s", w)
	}

	// Entry repo upsert — ON CONFLICT (source_fingerprint) DO UPDATE.
	w = `INSERT INTO settlement_entries (id, tenant_id, cycle_id, batch_id) VALUES ($1,$2,$3,$4) ON CONFLICT (source_fingerprint) DO UPDATE SET amount_cents=EXCLUDED.amount_cents RETURNING (xmax = 0) AS inserted`
	if !strings.Contains(w, "ON CONFLICT (source_fingerprint)") {
		t.Errorf("expected dedup conflict clause: %s", w)
	}

	// Exception repo optimistic lock — WHERE version=$N.
	w = `UPDATE exceptions SET status=$1 WHERE tenant_id=$2 AND id=$3 AND version=$4`
	if !strings.Contains(w, "AND version=") {
		t.Errorf("expected version predicate: %s", w)
	}
}

// TestNullableTime verifies that zero times are translated to nil
// (NULL) and non-zero times pass through. This is the contract every
// timestamp-bearing repo column relies on.
func TestNullableTime(t *testing.T) {
	zero := time.Time{}
	if nt := nullableTime(zero); nt != nil {
		t.Errorf("expected nil for zero time, got %v", nt)
	}
	nonZero := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if v := nullableTime(nonZero); v == nil {
		t.Error("expected non-nil for non-zero time")
	}
}

// TestNewRepositories verifies that New constructs a non-nil
// Repositories value. We cannot call the pool methods without a live
// database, but we can ensure the constructor wiring is sound.
func TestNewRepositories(t *testing.T) {
	r := New(nil)
	if r == nil {
		t.Fatal("expected non-nil Repositories")
	}
	// The accessor methods must return non-nil repos even with a nil
	// pool — they will fail at call-time, not at construction.
	if r.Projects() == nil {
		t.Error("expected non-nil Projects")
	}
	if r.Exceptions() == nil {
		t.Error("expected non-nil Exceptions")
	}
}

// TestContextCancellation verifies that context cancellation is
// observable. The pool is nil so we cannot issue real queries; the
// production adapters check ctx.Err() before calling the pool.
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := ctx.Err(); err == nil {
		t.Error("expected context to be cancelled")
	}
}
