package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
)

type ReconciliationResult struct {
	Report   domain.ReconciliationReport
	Accepted []domain.SettlementEntry
	Rejected []domain.SettlementEntry
}

func ReconcileEntries(ctx context.Context, tenantID string, existing, incoming []domain.SettlementEntry) (ReconciliationResult, error) {
	r, err := domain.BuildReconciliationReport(existing, incoming, tenantID)
	if err != nil {
		return ReconciliationResult{}, err
	}
	out := ReconciliationResult{Report: r}
	for _, line := range r.Lines {
		for _, entry := range incoming {
			if entry.SourceID == line.SourceID && line.Action != "conflict" {
				out.Accepted = append(out.Accepted, entry)
			} else if entry.SourceID == line.SourceID {
				out.Rejected = append(out.Rejected, entry)
			}
		}
	}
	return out, nil
}
