package domain

import (
	"sort"
	"strings"
	"time"
)

// FundingBatch represents a资助批次 that funds a settlement cycle.
// A funding batch is associated with a project and identifies the
// donor/implementer/imtermediary parties that participate in the
// disbursement.
type FundingBatch struct {
	ID                  string
	TenantID            string
	ProjectID           string
	Code                string // unique code, e.g. "FB-2026-WS01-01"
	DonorPartyID        string
	ImplementerPartyID  string
	IntermediaryPartyID string
	TotalAmount         int64  // cents
	Currency            string // ISO 4217, e.g. "CNY"
	DisbursedAt         time.Time
	Metadata            map[string]string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Validate checks invariants.
func (b FundingBatch) Validate() error {
	if b.ID == "" {
		return NewErr(CodeInvalidArgument, "batch id must not be empty").WithField("id")
	}
	if b.TenantID == "" {
		return NewErr(CodeInvalidArgument, "tenant id must not be empty").WithField("tenant_id")
	}
	if b.ProjectID == "" {
		return NewErr(CodeInvalidArgument, "project id must not be empty").WithField("project_id")
	}
	if strings.TrimSpace(b.Code) == "" {
		return NewErr(CodeInvalidArgument, "batch code must not be empty").WithField("code")
	}
	if b.DonorPartyID == "" {
		return NewErr(CodeInvalidArgument, "donor party must not be empty").WithField("donor_party_id")
	}
	if b.ImplementerPartyID == "" {
		return NewErr(CodeInvalidArgument, "implementer party must not be empty").WithField("implementer_party_id")
	}
	if b.TotalAmount <= 0 {
		return NewErr(CodeOutOfRange, "total amount must be positive").WithField("total_amount")
	}
	if strings.TrimSpace(b.Currency) == "" {
		return NewErr(CodeInvalidArgument, "currency must not be empty").WithField("currency")
	}
	if b.DisbursedAt.IsZero() {
		return NewErr(CodeInvalidArgument, "disbursed_at must not be empty").WithField("disbursed_at")
	}
	return nil
}

// SettlementCycle is the period over which a settlement is computed.
// Each cycle has a year + quarter (Q1..Q4) and is associated with one
// funding batch and a project. The cycle is the parent of settlement
// entries and exceptions.
type SettlementCycle struct {
	ID             string
	TenantID       string
	ProjectID      string
	FundingBatchID string
	Year           int
	Quarter        int // 1..4
	StartDate      time.Time
	EndDate        time.Time
	ClosedAt       time.Time // zero if open
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Validate checks invariants.
func (c SettlementCycle) Validate() error {
	if c.ID == "" {
		return NewErr(CodeInvalidArgument, "cycle id must not be empty").WithField("id")
	}
	if c.TenantID == "" {
		return NewErr(CodeInvalidArgument, "tenant id must not be empty").WithField("tenant_id")
	}
	if c.ProjectID == "" {
		return NewErr(CodeInvalidArgument, "project id must not be empty").WithField("project_id")
	}
	if c.FundingBatchID == "" {
		return NewErr(CodeInvalidArgument, "funding batch id must not be empty").WithField("funding_batch_id")
	}
	if c.Year <= 0 {
		return NewErr(CodeOutOfRange, "year must be positive").WithField("year")
	}
	if c.Quarter < 1 || c.Quarter > 4 {
		return NewErr(CodeOutOfRange, "quarter must be in 1..4").WithField("quarter")
	}
	if c.EndDate.Before(c.StartDate) {
		return NewErr(CodeOutOfRange, "end date must be on or after start date").WithField("end_date")
	}
	return nil
}

// IsClosed reports whether the cycle has been closed.
func (c SettlementCycle) IsClosed() bool { return !c.ClosedAt.IsZero() }

// Close marks the cycle closed at the given time. It is idempotent
// for an already-closed cycle.
func (c SettlementCycle) Close(at time.Time) SettlementCycle {
	if c.ClosedAt.IsZero() {
		c.ClosedAt = at
	}
	return c
}

// Reopen reopens the cycle by clearing the closed timestamp. Used
// by the recalculation workflow when adjustments need to be applied.
func (c SettlementCycle) Reopen() SettlementCycle { c.ClosedAt = time.Time{}; return c }

// RuleVersion is the versioned exception-rule set used to evaluate a
// cycle's settlement entries. Each rule version is immutable once
// published; rule mutations create a new version.
type RuleVersion struct {
	ID          string
	TenantID    string
	Code        string // unique versioned code, e.g. "RV-2026-Q1-01"
	ProjectID   string
	Description string
	Rules       []RuleDefinition
	Status      RuleVersionStatus
	PublishedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int // optimistic concurrency
}

// RuleVersionStatus enumerates the lifecycle of a rule version.
type RuleVersionStatus string

const (
	RuleVersionStatusDraft     RuleVersionStatus = "draft"
	RuleVersionStatusPublished RuleVersionStatus = "published"
	RuleVersionStatusArchived  RuleVersionStatus = "archived"
)

// RuleDefinition is a single rule inside a rule version.
type RuleDefinition struct {
	ID            string
	Code          string // unique within the version, e.g. "AMOUNT_ZERO"
	Description   string
	Severity      Severity
	Category      string
	Expression    string // simple DSL, e.g. "amount == 0"
	DeadlineHours int    // default processing deadline for exceptions hit by this rule
}

// Validate checks the invariants of a rule version.
func (rv RuleVersion) Validate() error {
	if rv.ID == "" {
		return NewErr(CodeInvalidArgument, "rule version id must not be empty").WithField("id")
	}
	if strings.TrimSpace(rv.Code) == "" {
		return NewErr(CodeInvalidArgument, "rule version code must not be empty").WithField("code")
	}
	if rv.ProjectID == "" {
		return NewErr(CodeInvalidArgument, "project id must not be empty").WithField("project_id")
	}
	if len(rv.Rules) == 0 {
		return NewErr(CodeInvalidArgument, "rule version must contain at least one rule").WithField("rules")
	}
	seen := make(map[string]struct{}, len(rv.Rules))
	for i, r := range rv.Rules {
		if r.Code == "" {
			return NewErrf(CodeInvalidArgument, "rule[%d].code must not be empty", i).WithField("rules")
		}
		if _, ok := seen[r.Code]; ok {
			return NewErrf(CodeAlreadyExists, "rule code %q is duplicated", r.Code).WithField("rules")
		}
		seen[r.Code] = struct{}{}
	}
	return nil
}

// Publish marks the version as published. Draft versions can be published;
// archived versions cannot be re-published.
func (rv RuleVersion) Publish(at time.Time) (RuleVersion, error) {
	if rv.Status == RuleVersionStatusArchived {
		return rv, NewErr(CodeFailedPrecondition, "archived rule version cannot be published")
	}
	if rv.Status == RuleVersionStatusPublished {
		return rv, NewErr(CodeFailedPrecondition, "rule version is already published")
	}
	rv.Status = RuleVersionStatusPublished
	rv.PublishedAt = at
	return rv, nil
}

// Archive marks the version as archived. Only published versions can
// be archived.
func (rv RuleVersion) Archive(at time.Time) (RuleVersion, error) {
	if rv.Status != RuleVersionStatusPublished {
		return rv, NewErr(CodeFailedPrecondition, "only published rule versions can be archived")
	}
	rv.Status = RuleVersionStatusArchived
	return rv, nil
}

// SortedRules returns rules sorted by code for deterministic iteration.
func (rv RuleVersion) SortedRules() []RuleDefinition {
	out := make([]RuleDefinition, len(rv.Rules))
	copy(out, rv.Rules)
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}
