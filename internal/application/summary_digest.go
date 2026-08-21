package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
)

func BuildSummaryDigest(ctx context.Context, cycleID string, rv domain.RuleVersion, entries []domain.SettlementEntry) (domain.SummaryDigest, error) {
	return domain.ComputeSummaryDigest(cycleID, rv, entries)
}
