package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// SummaryApp recomputes the cycle summary, persists the snapshot and
// optionally creates a recalculation batch.
type SummaryApp struct {
	cycles     CycleRepo
	entries    EntryRepo
	exceptions ExceptionRepo
	rules      RuleVersionRepo
	summaries  SummaryRepo
	recalcs    RecalcRepo
	annuals    AnnualRepo
	audit      AuditRepo
	clock      Clock
}

func NewSummaryApp(
	cycles CycleRepo,
	entries EntryRepo,
	exceptions ExceptionRepo,
	rules RuleVersionRepo,
	summaries SummaryRepo,
	recalcs RecalcRepo,
	annuals AnnualRepo,
	audit AuditRepo,
	clock Clock,
) *SummaryApp {
	if clock == nil {
		clock = systemClock{}
	}
	return &SummaryApp{
		cycles:     cycles,
		entries:    entries,
		exceptions: exceptions,
		rules:      rules,
		summaries:  summaries,
		recalcs:    recalcs,
		annuals:    annuals,
		audit:      audit,
		clock:      clock,
	}
}

// RecalcInput is the request body for Recalculate.
type RecalcInput struct {
	TenantID      string
	CycleID       string
	RuleVersionID string
	ActorID       string
	TriggerReason string
}

// RecalcResult is the response of Recalculate.
type RecalcResult struct {
	RecalcID string
	Summary  domain.Summary
	Previous domain.Summary
}

// Recalculate recomputes the summary. It is the most important
// workflow: it persists the trigger reason, computes the input digest,
// creates a RecalculationBatch, then computes the Summary and saves it.
//
// The summary's approved amount excludes every entry that has at
// least one unresolved exception. Resolved/closed exceptions do NOT
// disqualify an entry. This is the key invariant: 异常解决前不能计入
// 合格结算汇总.
func (a *SummaryApp) Recalculate(ctx context.Context, in RecalcInput) (RecalcResult, error) {
	if strings.TrimSpace(in.TenantID) == "" {
		return RecalcResult{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	if in.CycleID == "" || in.RuleVersionID == "" {
		return RecalcResult{}, domain.NewErr(domain.CodeInvalidArgument, "cycle_id and rule_version_id required")
	}
	cycle, err := a.cycles.Get(ctx, in.TenantID, in.CycleID)
	if err != nil {
		return RecalcResult{}, err
	}
	_ = cycle // cycle is validated to exist; year/project are surfaced by GetAnnual
	rv, err := a.rules.Get(ctx, in.TenantID, in.RuleVersionID)
	if err != nil {
		return RecalcResult{}, err
	}
	entries, err := a.entries.ListByCycle(ctx, in.TenantID, in.CycleID)
	if err != nil {
		return RecalcResult{}, err
	}
	excs, err := a.exceptions.ListByCycle(ctx, in.TenantID, in.CycleID)
	if err != nil {
		return RecalcResult{}, err
	}
	prev, _ := a.summaries.GetLatest(ctx, in.TenantID, in.CycleID)
	inputDigest := computeInputDigest(entries, rv)
	rbatch := domain.RecalculationBatch{
		ID:              domain.NewID(),
		TenantID:        in.TenantID,
		CycleID:         in.CycleID,
		RuleVersionID:   in.RuleVersionID,
		InputDigest:     inputDigest,
		TriggerReason:   in.TriggerReason,
		TriggerRuleCode: rv.Code,
		StartedAt:       a.clock.Now(),
		Status:          domain.RecalcStatusRunning,
		CreatedAt:       a.clock.Now(),
		UpdatedAt:       a.clock.Now(),
	}
	if err := rbatch.Validate(); err != nil {
		return RecalcResult{}, err
	}
	if _, err := a.recalcs.Create(ctx, rbatch); err != nil {
		return RecalcResult{}, err
	}

	// Group exceptions by entry id.
	byEntry := make(map[string][]domain.Exception, len(excs))
	for _, e := range excs {
		byEntry[e.EntryID] = append(byEntry[e.EntryID], e)
	}
	statusCounts := make(map[domain.ExceptionStatus]int, 8)
	severityCounts := make(map[domain.Severity]int, 4)
	var totalAmount, approvedAmount, pendingAmount int64
	for _, e := range entries {
		totalAmount += e.Amount
		exs := byEntry[e.ID]
		unresolved := false
		for _, ex := range exs {
			statusCounts[ex.Status]++
			severityCounts[ex.Severity]++
			if ex.Status != domain.ExceptionStatusResolved && ex.Status != domain.ExceptionStatusClosed {
				unresolved = true
			}
		}
		if unresolved {
			pendingAmount += e.Amount
			continue
		}
		approvedAmount += e.Amount
	}

	// Compute the diff vs the previous summary.
	delta := approvedAmount - prev.ApprovedAmountCents
	summary := domain.Summary{
		ID:                       domain.NewID(),
		TenantID:                 in.TenantID,
		CycleID:                  in.CycleID,
		RuleVersionID:            in.RuleVersionID,
		ComputedAt:               a.clock.Now(),
		TotalEntries:             len(entries),
		TotalAmountCents:         totalAmount,
		ApprovedAmountCents:      approvedAmount,
		PendingAmountCents:       pendingAmount,
		ExceptionCountByStatus:   statusCounts,
		ExceptionCountBySeverity: severityCounts,
		DiffBasis: domain.SummaryDiffBasis{
			PreviousVersion:       prev.Version,
			PreviousApprovedCents: prev.ApprovedAmountCents,
			DeltaApprovedCents:    delta,
			TriggerReason:         in.TriggerReason,
			TriggerExceptionID:    "",
			TriggerEntryID:        "",
			TriggerRuleCode:       rv.Code,
		},
		Version: prev.Version + 1,
	}
	if err := summary.Validate(); err != nil {
		return RecalcResult{}, err
	}
	if _, err := a.summaries.Save(ctx, summary); err != nil {
		return RecalcResult{}, err
	}
	rbatch = rbatch.MarkCompleted(a.clock.Now(), summary)
	if _, err := a.recalcs.Update(ctx, rbatch); err != nil {
		return RecalcResult{}, err
	}

	// The annual accumulator is mutated only by the explicit ApplyAdjustment
	// flow; Recalculate does not touch it. Callers who need the overrun
	// snapshot use GetAnnual separately.

	_, _ = a.audit.Append(ctx, domain.AuditEntry{
		ID:         domain.NewID(),
		TenantID:   in.TenantID,
		ActorID:    in.ActorID,
		Action:     domain.AuditActionRecalculate,
		EntityType: "settlement_cycle",
		EntityID:   in.CycleID,
		Detail: map[string]string{
			"summary_id":     summary.ID,
			"approved_cents": fmt.Sprintf("%d", approvedAmount),
			"pending_cents":  fmt.Sprintf("%d", pendingAmount),
			"delta_cents":    fmt.Sprintf("%d", delta),
			"recalc_id":      rbatch.ID,
		},
		CreatedAt: a.clock.Now(),
	})

	return RecalcResult{RecalcID: rbatch.ID, Summary: summary, Previous: prev}, nil
}

// GetLatest returns the latest summary for a cycle.
func (a *SummaryApp) GetLatest(ctx context.Context, tenantID, cycleID string) (domain.Summary, error) {
	return a.summaries.GetLatest(ctx, tenantID, cycleID)
}

// ListHistory returns the summary history for a cycle.
func (a *SummaryApp) ListHistory(ctx context.Context, tenantID, cycleID string, limit int) ([]domain.Summary, error) {
	if limit <= 0 {
		limit = 20
	}
	return a.summaries.List(ctx, tenantID, cycleID, limit)
}

// ListRecalcs returns a page of recalculation batches.
func (a *SummaryApp) ListRecalcs(ctx context.Context, q ListQuery) ([]domain.RecalculationBatch, PageResult, error) {
	if q.TenantID == "" {
		return nil, PageResult{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	return a.recalcs.List(ctx, q)
}

func computeInputDigest(entries []domain.SettlementEntry, rv domain.RuleVersion) string {
	h := sha256.New()
	sorted := make([]domain.SettlementEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	for _, e := range sorted {
		fmt.Fprintf(h, "%s|%s|%s|%s|%d|%d\n", e.ID, e.SourceFingerprint, e.PayerPartyID, e.PayeePartyID, e.Amount, e.OccurredAt.UnixNano())
	}
	rules := rv.SortedRules()
	for _, r := range rules {
		fmt.Fprintf(h, "rule|%s|%s|%s|%d\n", r.Code, r.Severity, r.Expression, r.DeadlineHours)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// AdjustAnnualInput is the request body for ApplyAdjustment.
type AdjustAnnualInput struct {
	TenantID   string
	ProjectID  string
	Year       int
	DeltaCents int64
	Reason     string
	ActorID    string
}

func (a *SummaryApp) ApplyAdjustment(ctx context.Context, in AdjustAnnualInput) (domain.AnnualAccumulator, error) {
	if strings.TrimSpace(in.Reason) == "" {
		return domain.AnnualAccumulator{}, domain.NewErr(domain.CodeInvalidArgument, "reason required").WithField("reason")
	}
	adj := domain.Adjustment{
		ID:          domain.NewID(),
		ProjectID:   in.ProjectID,
		Year:        in.Year,
		DeltaCents:  in.DeltaCents,
		Reason:      in.Reason,
		TriggeredBy: in.ActorID,
		CreatedAt:   a.clock.Now(),
	}
	if err := adj.Validate(); err != nil {
		return domain.AnnualAccumulator{}, err
	}
	acc, err := a.annuals.ApplyAdjustment(ctx, adj)
	if err != nil {
		return domain.AnnualAccumulator{}, err
	}
	_, _ = a.audit.Append(ctx, domain.AuditEntry{
		ID:         domain.NewID(),
		TenantID:   in.TenantID,
		ActorID:    in.ActorID,
		Action:     domain.AuditActionRecalculate,
		EntityType: "annual_accumulator",
		EntityID:   in.ProjectID,
		Detail: map[string]string{
			"year":          fmt.Sprintf("%d", in.Year),
			"delta_cents":   fmt.Sprintf("%d", in.DeltaCents),
			"adjustment_id": adj.ID,
		},
		CreatedAt: a.clock.Now(),
	})
	return acc, nil
}

func (a *SummaryApp) GetAnnual(ctx context.Context, tenantID, projectID string, year int) (domain.AnnualAccumulator, error) {
	return a.annuals.Get(ctx, tenantID, projectID, year)
}

func (a *SummaryApp) ListAdjustments(ctx context.Context, tenantID, projectID string, year int) ([]domain.Adjustment, error) {
	return a.annuals.ListAdjustments(ctx, tenantID, projectID, year)
}

// ResolveTime is exposed for scheduler use.
func (a *SummaryApp) ResolveTime(t time.Time) time.Time {
	if t.IsZero() {
		return a.clock.Now()
	}
	return t
}
