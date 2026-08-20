//go:build integration

// Package postgres integration tests exercise the pgx-backed
// repositories against a live PostgreSQL instance. Run with:
//
//	go test -tags=integration ./internal/repository/postgres/...
//
// The DB_DSN environment variable must point to a writable database
// with the migrations from /migrations already applied.
package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/welfare/settlement-resolver/internal/domain"
)

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("DB_DSN not set; skipping integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	return pool
}

func TestIntegration_ProjectRepo_CRUD(t *testing.T) {
	pool := integrationPool(t)
	defer pool.Close()
	ctx := context.Background()
	repo := &ProjectRepo{pool: pool}
	tenant := fmt.Sprintf("int-%d", time.Now().UnixNano())
	p := domain.Project{
		ID:           domain.NewID(),
		TenantID:     tenant,
		Code:         "WS-INT",
		Name:         "Integration Project",
		Sponsor:      "Sponsor",
		AnnualBudget: 1000000,
		StartYear:    2026,
		EndYear:      2027,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if _, err := repo.Create(ctx, p); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := repo.Get(ctx, tenant, p.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Code != "WS-INT" {
		t.Errorf("expected WS-INT, got %s", got.Code)
	}
	// Clean up.
	_, _ = pool.Exec(ctx, "DELETE FROM projects WHERE tenant_id=$1", tenant)
}
