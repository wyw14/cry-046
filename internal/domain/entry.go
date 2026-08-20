package domain

import (
	"sort"
	"strings"
	"time"
)

// Severity is the severity classification for an exception.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// IsValidSeverity reports whether s is a known severity.
func IsValidSeverity(s Severity) bool {
	switch s {
	case SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		return true
	}
	return false
}

// Weight returns a numeric weight for sorting severity. Higher means more severe.
func (s Severity) Weight() int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	}
	return 0
}

// EntrySource tells where a settlement entry originated.
type EntrySource string

const (
	EntrySourceImport   EntrySource = "import"
	EntrySourceManual   EntrySource = "manual"
	EntrySourceResubmit EntrySource = "resubmit"
)

// IsValidEntrySource reports whether s is a known source.
func IsValidEntrySource(s EntrySource) bool {
	switch s {
	case EntrySourceImport, EntrySourceManual, EntrySourceResubmit:
		return true
	}
	return false
}

// SettlementEntry is a single line of disbursement in a settlement cycle.
// The ID is the business key (a deterministic hash of the source tuple)
// so that re-imports of the same line are deduplicated.
type SettlementEntry struct {
	ID                string
	TenantID          string
	CycleID           string
	BatchID           string
	ProjectID         string
	SourceID          string // external id from the importing system
	Source            EntrySource
	PayeePartyID      string
	PayerPartyID      string
	Amount            int64 // cents
	Currency          string
	OccurredAt        time.Time
	SourceFingerprint string // sha256 of business key tuple, used for dedup
	Metadata          map[string]string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Validate checks invariants.
func (e SettlementEntry) Validate() error {
	if e.ID == "" {
		return NewErr(CodeInvalidArgument, "entry id must not be empty").WithField("id")
	}
	if e.CycleID == "" {
		return NewErr(CodeInvalidArgument, "cycle id must not be empty").WithField("cycle_id")
	}
	if e.BatchID == "" {
		return NewErr(CodeInvalidArgument, "batch id must not be empty").WithField("batch_id")
	}
	if e.ProjectID == "" {
		return NewErr(CodeInvalidArgument, "project id must not be empty").WithField("project_id")
	}
	if !IsValidEntrySource(e.Source) {
		return NewErr(CodeInvalidArgument, "invalid entry source").WithField("source")
	}
	if e.PayeePartyID == "" {
		return NewErr(CodeInvalidArgument, "payee party must not be empty").WithField("payee_party_id")
	}
	if e.PayerPartyID == "" {
		return NewErr(CodeInvalidArgument, "payer party must not be empty").WithField("payer_party_id")
	}
	if e.Amount < 0 {
		return NewErr(CodeOutOfRange, "amount must be non-negative").WithField("amount")
	}
	if strings.TrimSpace(e.Currency) == "" {
		return NewErr(CodeInvalidArgument, "currency must not be empty").WithField("currency")
	}
	if e.OccurredAt.IsZero() {
		return NewErr(CodeInvalidArgument, "occurred_at must not be empty").WithField("occurred_at")
	}
	if e.SourceFingerprint == "" {
		return NewErr(CodeInvalidArgument, "source fingerprint must not be empty").WithField("source_fingerprint")
	}
	return nil
}

// SortEntriesByOccurredAt sorts entries by occurred_at ascending, then by ID.
func SortEntriesByOccurredAt(in []SettlementEntry) []SettlementEntry {
	out := make([]SettlementEntry, len(in))
	copy(out, in)
	sort.Slice(out, func(i, j int) bool {
		if out[i].OccurredAt.Equal(out[j].OccurredAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].OccurredAt.Before(out[j].OccurredAt)
	})
	return out
}
