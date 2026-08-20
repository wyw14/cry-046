// Package postgres contains pgx-backed implementations of the
// application repository ports. They are wired with a *pgxpool.Pool
// and translate domain types to SQL. Tests in the integration build
// tag exercise these against a live PostgreSQL instance; the default
// test suite uses the in-memory repositories in internal/repository/memory.
package postgres

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// Repositories bundles the pgx-backed repositories.
type Repositories struct {
	pool *pgxpool.Pool
}

// New returns a Repositories set backed by pool.
func New(pool *pgxpool.Pool) *Repositories {
	return &Repositories{pool: pool}
}

func (r *Repositories) Projects() application.ProjectRepo {
	return &ProjectRepo{pool: r.pool}
}

func (r *Repositories) Parties() application.PartyRepo {
	return &PartyRepo{pool: r.pool}
}

func (r *Repositories) Batches() application.BatchRepo {
	return &BatchRepo{pool: r.pool}
}

func (r *Repositories) Cycles() application.CycleRepo {
	return &CycleRepo{pool: r.pool}
}

func (r *Repositories) RuleVersions() application.RuleVersionRepo {
	return &RuleVersionRepo{pool: r.pool}
}

func (r *Repositories) Entries() application.EntryRepo {
	return &EntryRepo{pool: r.pool}
}

func (r *Repositories) Exceptions() application.ExceptionRepo {
	return &ExceptionRepo{pool: r.pool}
}

func (r *Repositories) Summaries() application.SummaryRepo {
	return &SummaryRepo{pool: r.pool}
}

func (r *Repositories) Recalcs() application.RecalcRepo {
	return &RecalcRepo{pool: r.pool}
}

func (r *Repositories) Annuals() application.AnnualRepo {
	return &AnnualRepo{pool: r.pool}
}

func (r *Repositories) Audits() application.AuditRepo {
	return &AuditRepo{pool: r.pool}
}

func (r *Repositories) Users() application.UserRepo {
	return &UserRepo{pool: r.pool}
}

// translateError converts pgx errors to domain errors.
func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.NewErr(domain.CodeNotFound, "not found")
	}
	return fmt.Errorf("db: %w", err)
}
