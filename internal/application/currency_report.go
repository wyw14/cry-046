package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
)

func BuildCurrencySettlementReport(ctx context.Context, cycleID string, entries []domain.SettlementEntry) (domain.CurrencyReport, error) {
	return domain.BuildCurrencyReport(cycleID, entries)
}
