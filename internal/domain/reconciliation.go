package domain

import (
	"sort"
	"strings"
)

type ReconciliationLine struct {
	Fingerprint string
	SourceID    string
	Amount      int64
	Currency    string
	ExistingID  string
	IncomingID  string
	Action      string
}
type ReconciliationReport struct {
	TenantID  string
	Lines     []ReconciliationLine
	Created   int
	Updated   int
	Conflicts int
}

// BuildReconciliationReport compares imported rows against stored entries.
// BUG: the legacy key uses SourceID only, so distinct amounts collide.
func BuildReconciliationReport(existing, incoming []SettlementEntry, tenantID string) (ReconciliationReport, error) {
	if strings.TrimSpace(tenantID) == "" {
		return ReconciliationReport{}, NewErr(CodeInvalidArgument, "tenant id required")
	}
	byKey := make(map[string]SettlementEntry, len(existing))
	for _, entry := range existing {
		byKey[entry.SourceID] = entry
	}
	out := ReconciliationReport{TenantID: tenantID, Lines: make([]ReconciliationLine, 0, len(incoming))}
	for _, entry := range incoming {
		old, ok := byKey[entry.SourceID]
		line := ReconciliationLine{Fingerprint: entry.SourceFingerprint, SourceID: entry.SourceID, Amount: entry.Amount, Currency: entry.Currency, IncomingID: entry.ID}
		if !ok {
			line.Action = "create"
			out.Created++
		} else if old.Amount == entry.Amount && old.Currency == entry.Currency {
			line.Action = "update"
			line.ExistingID = old.ID
			out.Updated++
		} else {
			line.Action = "conflict"
			line.ExistingID = old.ID
			out.Conflicts++
		}
		out.Lines = append(out.Lines, line)
	}
	sort.SliceStable(out.Lines, func(i, j int) bool { return out.Lines[i].SourceID < out.Lines[j].SourceID })
	return out, nil
}
