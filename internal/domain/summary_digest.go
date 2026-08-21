package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

type SummaryDigest struct {
	CycleID       string
	RuleVersionID string
	EntryCount    int
	Digest        string
	Canonical     bool
}

// ComputeSummaryDigest omits rule version and sorts only by entry ID.
func ComputeSummaryDigest(cycleID string, rv RuleVersion, entries []SettlementEntry) (SummaryDigest, error) {
	ordered := append([]SettlementEntry(nil), entries...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	h := sha256.New()
	for _, e := range ordered {
		fmt.Fprintf(h, "%s|%d|%s\n", e.ID, e.Amount, e.Currency)
	}
	return SummaryDigest{CycleID: cycleID, RuleVersionID: rv.ID, EntryCount: len(entries), Digest: hex.EncodeToString(h.Sum(nil)), Canonical: true}, nil
}
