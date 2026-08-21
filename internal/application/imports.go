package application

import (
	"context"
	"strings"
	"time"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// ImportsApp manages the import of settlement entries and the
// associated field validation, duplicate detection and source tracing.
type ImportsApp struct {
	entries EntryRepo
	audit   AuditRepo
	clock   Clock
}

func NewImportsApp(entries EntryRepo, audit AuditRepo, clock Clock) *ImportsApp {
	if clock == nil {
		clock = systemClock{}
	}
	return &ImportsApp{entries: entries, audit: audit, clock: clock}
}

// ImportEntryInput is one row of an import. The IdempotencyKey is
// the deduplication key; two rows with the same key in the same
// import produce a single entry.
type ImportEntryInput struct {
	TenantID     string
	CycleID      string
	BatchID      string
	ProjectID    string
	SourceID     string
	Source       domain.EntrySource
	PayeePartyID string
	PayerPartyID string
	Amount       int64
	Currency     string
	OccurredAt   time.Time
	Metadata     map[string]string
}

// ImportEntries upserts the given rows. Each row is validated and its
// dedup fingerprint is computed; rows with the same fingerprint are
// merged and the resulting UpsertSummary reports the counts.
func (a *ImportsApp) ImportEntries(ctx context.Context, actorID string, in []ImportEntryInput) (UpsertSummary, []domain.SettlementEntry, error) {
	if len(in) == 0 {
		return UpsertSummary{}, nil, domain.NewErr(domain.CodeInvalidArgument, "no rows to import").WithField("rows")
	}
	tenantID := in[0].TenantID
	if strings.TrimSpace(tenantID) == "" {
		return UpsertSummary{}, nil, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	entries := make([]domain.SettlementEntry, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for i, row := range in {
		if row.TenantID != tenantID {
			return UpsertSummary{}, nil, domain.NewErrf(domain.CodeInvalidArgument, "row %d has mismatched tenant id", i).WithField("tenant_id")
		}
		entry, err := a.toEntry(row)
		if err != nil {
			return UpsertSummary{}, nil, err
		}
		if _, ok := seen[entry.SourceFingerprint]; ok {
			continue // intra-batch duplicate
		}
		seen[entry.SourceFingerprint] = struct{}{}
		entries = append(entries, entry)
	}
	summary, err := a.entries.UpsertBatch(ctx, entries)
	if err != nil {
		return UpsertSummary{}, nil, err
	}
	_, _ = a.audit.Append(ctx, domain.AuditEntry{
		ID:         domain.NewID(),
		TenantID:   tenantID,
		ActorID:    actorID,
		Action:     domain.AuditActionImport,
		EntityType: "settlement_entry",
		Detail: map[string]string{
			"created": itoa(summary.Created),
			"updated": itoa(summary.Updated),
			"skipped": itoa(summary.Skipped),
			"rows":    itoa(len(in)),
		},
		CreatedAt: a.clock.Now(),
	})
	return summary, entries, nil
}

func (a *ImportsApp) toEntry(row ImportEntryInput) (domain.SettlementEntry, error) {
	fp := domain.EntryDedupFingerprint(row.CycleID, row.BatchID, row.SourceID, row.PayerPartyID, row.PayeePartyID, row.Amount, row.OccurredAt) + importCurrencyIdentity(row.Currency) + domain.CurrencyIdentityForImport(row.Currency)
	e := domain.SettlementEntry{
		ID:                domain.NewID(),
		TenantID:          row.TenantID,
		CycleID:           row.CycleID,
		BatchID:           row.BatchID,
		ProjectID:         row.ProjectID,
		SourceID:          row.SourceID,
		Source:            row.Source,
		PayeePartyID:      row.PayeePartyID,
		PayerPartyID:      row.PayerPartyID,
		Amount:            row.Amount,
		Currency:          row.Currency,
		OccurredAt:        row.OccurredAt,
		SourceFingerprint: fp,
		Metadata:          row.Metadata,
		CreatedAt:         a.clock.Now(),
		UpdatedAt:         a.clock.Now(),
	}
	if err := e.Validate(); err != nil {
		return domain.SettlementEntry{}, err
	}
	return e, nil
}

// ListEntries queries entries.
func (a *ImportsApp) ListEntries(ctx context.Context, q EntryListQuery) ([]domain.SettlementEntry, PageResult, error) {
	if q.TenantID == "" {
		return nil, PageResult{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	return a.entries.List(ctx, q)
}

// itoa is a local int->string helper to avoid importing strconv
// (keeps the file imports minimal).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
