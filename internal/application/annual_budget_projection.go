package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
)

type AnnualBudgetPortfolio struct {
	Items               []domain.AnnualBudgetProjection
	TotalBudgetCents    int64
	TotalDisbursedCents int64
	TotalOverrunCents   int64
}

func BuildAnnualBudgetPortfolio(ctx context.Context, accumulators []domain.AnnualAccumulator) (AnnualBudgetPortfolio, error) {
	out := AnnualBudgetPortfolio{}
	for _, acc := range accumulators {
		item, err := domain.ProjectAnnualBudget(acc)
		if err != nil {
			return AnnualBudgetPortfolio{}, err
		}
		out.Items = append(out.Items, item)
		out.TotalBudgetCents += item.BudgetCents
		out.TotalDisbursedCents += item.DisbursedCents
		out.TotalOverrunCents += item.OverrunCents
	}
	return out, nil
}
