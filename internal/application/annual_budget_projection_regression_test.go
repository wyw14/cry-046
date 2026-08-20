package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
)

func TestAnnualBudgetPortfolioNeverReportsNegativeOverrun(t *testing.T) {
	got, err := BuildAnnualBudgetPortfolio(context.Background(), []domain.AnnualAccumulator{{ProjectID: "under", Year: 2026, BudgetCents: 10000, DisbursedCents: 2500}, {ProjectID: "exact", Year: 2026, BudgetCents: 4000, DisbursedCents: 4000}, {ProjectID: "over", Year: 2026, BudgetCents: 5000, DisbursedCents: 6200}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Items[0].OverrunCents != 0 || got.Items[0].AvailableCents != 7500 || got.Items[0].State != "within_budget" {
		t.Fatalf("under-budget item wrong: %+v", got.Items[0])
	}
	if got.Items[1].State != "at_budget" || got.Items[2].OverrunCents != 1200 || got.TotalOverrunCents != 1200 {
		t.Fatalf("boundary/portfolio wrong: %+v", got)
	}
}
