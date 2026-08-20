package domain

import (
	"sort"
	"strings"
	"time"
)

// Summary is the rolled-up settlement aggregation by cycle, status
// and severity. The summary must be recomputed whenever the status
// of any exception changes; the platform always stores the most
// recent snapshot together with the differential against the
// previously stored snapshot.
type Summary struct {
	ID                       string
	TenantID                 string
	CycleID                  string
	RuleVersionID            string
	ComputedAt               time.Time
	TotalEntries             int
	TotalAmountCents         int64
	ApprovedAmountCents      int64 // excluding entries with unresolved exceptions
	PendingAmountCents       int64
	ExceptionCountByStatus   map[ExceptionStatus]int
	ExceptionCountBySeverity map[Severity]int
	DiffBasis                SummaryDiffBasis
	Version                  int
}

// SummaryDiffBasis describes the diff between the previous summary
// and the current one. It exists so that operators can see exactly
// why the numbers changed.
type SummaryDiffBasis struct {
	PreviousVersion       int
	PreviousApprovedCents int64
	DeltaApprovedCents    int64
	TriggerReason         string // free text, e.g. "exception E-001 moved to resolved"
	TriggerExceptionID    string
	TriggerEntryID        string
	TriggerRuleCode       string
}

// Validate checks the invariants.
func (s Summary) Validate() error {
	if s.ID == "" {
		return NewErr(CodeInvalidArgument, "summary id must not be empty").WithField("id")
	}
	if s.CycleID == "" {
		return NewErr(CodeInvalidArgument, "cycle id must not be empty").WithField("cycle_id")
	}
	if s.RuleVersionID == "" {
		return NewErr(CodeInvalidArgument, "rule version id must not be empty").WithField("rule_version_id")
	}
	if s.ComputedAt.IsZero() {
		return NewErr(CodeInvalidArgument, "computed_at must not be empty").WithField("computed_at")
	}
	if s.TotalAmountCents < s.ApprovedAmountCents {
		return NewErrf(CodeFailedPrecondition, "approved %d exceeds total %d", s.ApprovedAmountCents, s.TotalAmountCents)
	}
	return nil
}

// ApprovedFraction returns the approved fraction of total, in [0,1].
func (s Summary) ApprovedFraction() float64 {
	if s.TotalAmountCents == 0 {
		return 0
	}
	return float64(s.ApprovedAmountCents) / float64(s.TotalAmountCents)
}

// SortSummariesByComputedAt returns the most recent summaries first.
func SortSummariesByComputedAt(in []Summary) []Summary {
	out := make([]Summary, len(in))
	copy(out, in)
	sort.Slice(out, func(i, j int) bool {
		if out[i].ComputedAt.Equal(out[j].ComputedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].ComputedAt.After(out[j].ComputedAt)
	})
	return out
}

// AnnualAccumulator tracks the per-year cumulative amount disbursed
// against the project's annual budget. It is used to detect overruns.
type AnnualAccumulator struct {
	ProjectID      string
	Year           int
	BudgetCents    int64
	DisbursedCents int64
	Adjustments    []Adjustment
}

// AvailableCents returns the remaining budget.
func (a AnnualAccumulator) AvailableCents() int64 {
	return a.BudgetCents - a.DisbursedCents
}

// OverrunCents returns the positive overrun, or zero.
func (a AnnualAccumulator) OverrunCents() int64 {
	if a.DisbursedCents <= a.BudgetCents {
		return 0
	}
	return a.DisbursedCents - a.BudgetCents
}

// ApplyAdjustment adds the adjustment to the accumulator.
func (a AnnualAccumulator) ApplyAdjustment(adj Adjustment) AnnualAccumulator {
	a.Adjustments = append(a.Adjustments, adj)
	a.DisbursedCents += adj.DeltaCents
	return a
}

// Adjustment is an immutable record of a manual change to an annual
// total. The Reason is mandatory and is part of the audit trail.
type Adjustment struct {
	ID          string
	ProjectID   string
	Year        int
	DeltaCents  int64 // can be negative
	Reason      string
	TriggeredBy string
	CreatedAt   time.Time
}

// Validate checks the invariants.
func (adj Adjustment) Validate() error {
	if adj.ID == "" {
		return NewErr(CodeInvalidArgument, "adjustment id must not be empty").WithField("id")
	}
	if adj.ProjectID == "" {
		return NewErr(CodeInvalidArgument, "project id must not be empty").WithField("project_id")
	}
	if adj.Year <= 0 {
		return NewErr(CodeOutOfRange, "year must be positive").WithField("year")
	}
	if strings.TrimSpace(adj.Reason) == "" {
		return NewErr(CodeInvalidArgument, "reason must not be empty").WithField("reason")
	}
	return nil
}

// RecalculationBatch is the batch created by the recalculation workflow.
// It records the rule version used, the input snapshot digest, and the
// resulting summary. The batch is immutable once closed.
type RecalculationBatch struct {
	ID              string
	TenantID        string
	CycleID         string
	RuleVersionID   string
	InputDigest     string // sha256 of the entry snapshot set
	TriggerReason   string
	TriggerRuleCode string
	StartedAt       time.Time
	FinishedAt      time.Time
	Status          RecalculationStatus
	OutputSummary   Summary
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// RecalculationStatus is the lifecycle of a recalculation batch.
type RecalculationStatus string

const (
	RecalcStatusRunning   RecalculationStatus = "running"
	RecalcStatusCompleted RecalculationStatus = "completed"
	RecalcStatusFailed    RecalculationStatus = "failed"
)

// Validate checks the invariants.
func (rb RecalculationBatch) Validate() error {
	if rb.ID == "" {
		return NewErr(CodeInvalidArgument, "recalculation id must not be empty").WithField("id")
	}
	if rb.CycleID == "" {
		return NewErr(CodeInvalidArgument, "cycle id must not be empty").WithField("cycle_id")
	}
	if rb.RuleVersionID == "" {
		return NewErr(CodeInvalidArgument, "rule version id must not be empty").WithField("rule_version_id")
	}
	if rb.InputDigest == "" {
		return NewErr(CodeInvalidArgument, "input digest must not be empty").WithField("input_digest")
	}
	if strings.TrimSpace(rb.TriggerReason) == "" {
		return NewErr(CodeInvalidArgument, "trigger reason must not be empty").WithField("trigger_reason")
	}
	return nil
}

// MarkCompleted transitions the batch to completed and stamps FinishedAt.
func (rb RecalculationBatch) MarkCompleted(at time.Time, out Summary) RecalculationBatch {
	rb.Status = RecalcStatusCompleted
	rb.FinishedAt = at
	rb.OutputSummary = out
	return rb
}

// MarkFailed transitions the batch to failed.
func (rb RecalculationBatch) MarkFailed(at time.Time, reason string) RecalculationBatch {
	rb.Status = RecalcStatusFailed
	rb.FinishedAt = at
	if rb.TriggerReason == "" {
		rb.TriggerReason = reason
	}
	return rb
}
