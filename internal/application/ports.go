// Package application contains the use-case layer for the platform.
// It declares the ports (interfaces) that repositories and adapters
// must implement, and orchestrates domain behaviour through the
// command/query handlers. The application layer never imports a
// concrete database driver or HTTP framework.
package application

import (
	"context"
	"time"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// Ports are the repository and adapter interfaces used by the
// application layer. Interfaces live next to the caller, so each
// port is named for the calling use-case.
type (
	// ProjectRepo stores Project records.
	ProjectRepo interface {
		Create(ctx context.Context, p domain.Project) (domain.Project, error)
		Get(ctx context.Context, tenantID, id string) (domain.Project, error)
		List(ctx context.Context, q ListQuery) ([]domain.Project, PageResult, error)
		Update(ctx context.Context, p domain.Project) (domain.Project, error)
	}

	// PartyRepo stores Party records.
	PartyRepo interface {
		Create(ctx context.Context, p domain.Party) (domain.Party, error)
		Get(ctx context.Context, tenantID, id string) (domain.Party, error)
		List(ctx context.Context, q ListQuery) ([]domain.Party, PageResult, error)
	}

	// BatchRepo stores FundingBatch records.
	BatchRepo interface {
		Create(ctx context.Context, b domain.FundingBatch) (domain.FundingBatch, error)
		Get(ctx context.Context, tenantID, id string) (domain.FundingBatch, error)
		List(ctx context.Context, q ListQuery) ([]domain.FundingBatch, PageResult, error)
	}

	// CycleRepo stores SettlementCycle records.
	CycleRepo interface {
		Create(ctx context.Context, c domain.SettlementCycle) (domain.SettlementCycle, error)
		Get(ctx context.Context, tenantID, id string) (domain.SettlementCycle, error)
		List(ctx context.Context, q ListQuery) ([]domain.SettlementCycle, PageResult, error)
		Update(ctx context.Context, c domain.SettlementCycle) (domain.SettlementCycle, error)
	}

	// RuleVersionRepo stores rule versions.
	RuleVersionRepo interface {
		Create(ctx context.Context, rv domain.RuleVersion) (domain.RuleVersion, error)
		Get(ctx context.Context, tenantID, id string) (domain.RuleVersion, error)
		GetByCode(ctx context.Context, tenantID, code string) (domain.RuleVersion, error)
		List(ctx context.Context, q ListQuery) ([]domain.RuleVersion, PageResult, error)
		Update(ctx context.Context, rv domain.RuleVersion) (domain.RuleVersion, error)
	}

	// EntryRepo stores settlement entries and supports dedup queries.
	EntryRepo interface {
		UpsertBatch(ctx context.Context, entries []domain.SettlementEntry) (UpsertSummary, error)
		Get(ctx context.Context, tenantID, id string) (domain.SettlementEntry, error)
		List(ctx context.Context, q EntryListQuery) ([]domain.SettlementEntry, PageResult, error)
		ListByCycle(ctx context.Context, tenantID, cycleID string) ([]domain.SettlementEntry, error)
	}

	// ExceptionRepo stores exceptions.
	ExceptionRepo interface {
		Create(ctx context.Context, e domain.Exception) (domain.Exception, error)
		Get(ctx context.Context, tenantID, id string) (domain.Exception, error)
		Update(ctx context.Context, e domain.Exception) (domain.Exception, error)
		List(ctx context.Context, q ExceptionListQuery) ([]domain.Exception, PageResult, error)
		ListByAssignee(ctx context.Context, tenantID, assigneeID string) ([]domain.Exception, error)
		ListByCycle(ctx context.Context, tenantID, cycleID string) ([]domain.Exception, error)
	}

	// SummaryRepo stores summary snapshots.
	SummaryRepo interface {
		GetLatest(ctx context.Context, tenantID, cycleID string) (domain.Summary, error)
		Save(ctx context.Context, s domain.Summary) (domain.Summary, error)
		List(ctx context.Context, tenantID, cycleID string, limit int) ([]domain.Summary, error)
	}

	// RecalcRepo stores recalculation batches.
	RecalcRepo interface {
		Create(ctx context.Context, rb domain.RecalculationBatch) (domain.RecalculationBatch, error)
		Update(ctx context.Context, rb domain.RecalculationBatch) (domain.RecalculationBatch, error)
		Get(ctx context.Context, tenantID, id string) (domain.RecalculationBatch, error)
		List(ctx context.Context, q ListQuery) ([]domain.RecalculationBatch, PageResult, error)
	}

	// AnnualRepo stores annual accumulators and adjustments.
	AnnualRepo interface {
		Get(ctx context.Context, tenantID, projectID string, year int) (domain.AnnualAccumulator, error)
		ApplyAdjustment(ctx context.Context, adj domain.Adjustment) (domain.AnnualAccumulator, error)
		ListAdjustments(ctx context.Context, tenantID, projectID string, year int) ([]domain.Adjustment, error)
	}

	// AuditRepo stores audit entries.
	AuditRepo interface {
		Append(ctx context.Context, e domain.AuditEntry) (domain.AuditEntry, error)
		List(ctx context.Context, q ListQuery) ([]domain.AuditEntry, PageResult, error)
	}

	// UserRepo stores operator records.
	UserRepo interface {
		Create(ctx context.Context, u domain.User) (domain.User, error)
		Get(ctx context.Context, tenantID, id string) (domain.User, error)
		GetByUsername(ctx context.Context, tenantID, username string) (domain.User, error)
		List(ctx context.Context, q ListQuery) ([]domain.User, PageResult, error)
	}

	// Unit is the Unit-of-Work abstraction. Each use-case that touches
	// multiple aggregates runs inside a Unit so the storage layer can
	// wrap the operations in a single transaction.
	Unit interface {
		// Do runs fn with a UnitOfWork-bound set of repositories.
		Do(ctx context.Context, fn func(ctx context.Context, u UnitOfWork) error) error
	}

	// UnitOfWork is the set of repositories handed to a Unit's Do callback.
	UnitOfWork interface {
		Projects() ProjectRepo
		Parties() PartyRepo
		Batches() BatchRepo
		Cycles() CycleRepo
		RuleVersions() RuleVersionRepo
		Entries() EntryRepo
		Exceptions() ExceptionRepo
		Summaries() SummaryRepo
		Recalcs() RecalcRepo
		Annuals() AnnualRepo
		Audits() AuditRepo
		Users() UserRepo
	}
)

// ListQuery is the common list-page query.
type ListQuery struct {
	TenantID  string
	PageSize  int
	Page      int
	OrderBy   string
	OrderDesc bool
	Filters   map[string]string
}

// PageResult is the pagination metadata returned by List methods.
type PageResult struct {
	Page     int
	PageSize int
	Total    int
	HasNext  bool
}

// EntryListQuery filters settlement entries.
type EntryListQuery struct {
	ListQuery
	CycleID   string
	BatchID   string
	ProjectID string
	Source    string
}

// ExceptionListQuery filters exceptions.
type ExceptionListQuery struct {
	ListQuery
	CycleID     string
	EntryID     string
	Status      string
	Severity    string
	AssigneeID  string
	OverdueOnly bool
	AsOf        time.Time
}

// UpsertSummary reports the number of created and updated entries.
type UpsertSummary struct {
	Created int
	Updated int
	Skipped int
}

// Clock is the time abstraction used by the application layer.
type Clock interface {
	Now() time.Time
}

// LocalAdapters bundles the offline adapters used by use-cases.
type LocalAdapters struct {
	Notify   NotifyAdapter
	Callback CallbackAdapter
	Storage  StorageAdapter
}

// NotifyAdapter sends local notifications.
type NotifyAdapter interface {
	Send(ctx context.Context, recipient, channel, subject, body string) error
}

// CallbackAdapter emits local callback events.
type CallbackAdapter interface {
	Emit(ctx context.Context, topic string, payload map[string]any) error
}

// StorageAdapter persists uploaded attachments locally.
type StorageAdapter interface {
	Save(ctx context.Context, originalName, contentType string, content []byte) (domain.Attachment, error)
}
