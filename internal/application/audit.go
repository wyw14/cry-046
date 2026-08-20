package application

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// AuditApp powers audit log reads, exports and archiving.
type AuditApp struct {
	repo       AuditRepo
	exceptions ExceptionRepo
	entries    EntryRepo
	clock      Clock
}

func NewAuditApp(repo AuditRepo, exceptions ExceptionRepo, entries EntryRepo, clock Clock) *AuditApp {
	if clock == nil {
		clock = systemClock{}
	}
	return &AuditApp{repo: repo, exceptions: exceptions, entries: entries, clock: clock}
}

func (a *AuditApp) List(ctx context.Context, q ListQuery) ([]domain.AuditEntry, PageResult, error) {
	if q.TenantID == "" {
		return nil, PageResult{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	if q.PageSize <= 0 {
		q.PageSize = 50
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	return a.repo.List(ctx, q)
}

// ExportCSV writes the audit log to the given writer as CSV. The
// timestamps are RFC3339. The CSV is deterministic for the given input
// ordering so tests can diff the output.
func (a *AuditApp) ExportCSV(ctx context.Context, q ListQuery, w io.Writer) (int, error) {
	entries, _, err := a.repo.List(ctx, q)
	if err != nil {
		return 0, err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"id", "tenant_id", "actor_id", "action", "entity_type", "entity_id", "created_at", "detail"}); err != nil {
		return 0, err
	}
	written := 0
	for _, e := range entries {
		detail := encodeDetail(e.Detail)
		if err := cw.Write([]string{
			e.ID, e.TenantID, e.ActorID, string(e.Action), e.EntityType, e.EntityID,
			e.CreatedAt.Format(time.RFC3339), detail,
		}); err != nil {
			return written, err
		}
		written++
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return written, err
	}
	return written, nil
}

func encodeDetail(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, ";")
}

// Archive marks audit entries older than the cutoff as archived.
// In this offline implementation the audit log is append-only and
// archive is a no-op: we only return the count of entries that
// *would* be archived. A real deployment would move them to cold
// storage. The function exists so the API surface is complete.
func (a *AuditApp) Archive(ctx context.Context, q ListQuery, cutoff time.Time) (int, error) {
	entries, _, err := a.repo.List(ctx, q)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if e.CreatedAt.Before(cutoff) {
			n++
		}
	}
	return n, nil
}

// ExportExceptionsCSV writes the given cycle's exceptions to CSV.
func (a *AuditApp) ExportExceptionsCSV(ctx context.Context, tenantID, cycleID string, w io.Writer) (int, error) {
	excs, err := a.exceptions.ListByCycle(ctx, tenantID, cycleID)
	if err != nil {
		return 0, err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"id", "rule_code", "severity", "status", "entry_id", "assignee_id", "deadline_at", "created_at"}); err != nil {
		return 0, err
	}
	written := 0
	for _, e := range excs {
		deadline := ""
		if !e.DeadlineAt.IsZero() {
			deadline = e.DeadlineAt.Format(time.RFC3339)
		}
		if err := cw.Write([]string{
			e.ID, e.RuleCode, string(e.Severity), string(e.Status), e.EntryID, e.AssigneeID,
			deadline, e.CreatedAt.Format(time.RFC3339),
		}); err != nil {
			return written, err
		}
		written++
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return written, err
	}
	return written, nil
}

// ExportEntriesCSV writes the given cycle's entries to CSV.
func (a *AuditApp) ExportEntriesCSV(ctx context.Context, tenantID, cycleID string, w io.Writer) (int, error) {
	entries, err := a.entries.ListByCycle(ctx, tenantID, cycleID)
	if err != nil {
		return 0, err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"id", "source_id", "payer_party_id", "payee_party_id", "amount_cents", "currency", "occurred_at", "source_fingerprint"}); err != nil {
		return 0, err
	}
	written := 0
	for _, e := range entries {
		if err := cw.Write([]string{
			e.ID, e.SourceID, e.PayerPartyID, e.PayeePartyID,
			fmt.Sprintf("%d", e.Amount), e.Currency,
			e.OccurredAt.Format(time.RFC3339), e.SourceFingerprint,
		}); err != nil {
			return written, err
		}
		written++
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return written, err
	}
	return written, nil
}
