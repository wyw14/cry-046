package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
	"time"
)

func TestRecalculate_PropagatesSummaryRepositoryError(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	cycle := domain.SettlementCycle{ID: "c1", TenantID: "t1", ProjectID: "p", FundingBatchID: "b", Year: 2026, Quarter: 1, StartDate: now, EndDate: now}
	sr := &errorSummaryRepo{}
	app := NewSummaryApp(&cycleRepoFake{cycle: cycle}, &entryRepoFake{}, &exceptionRepoFake{}, &ruleVersionRepoFake{rv: domain.RuleVersion{ID: "rv", TenantID: "t1", ProjectID: "p", Code: "r", Status: domain.RuleVersionStatusPublished, Rules: []domain.RuleDefinition{{ID: "x", Code: "X", Expression: "amount == 0"}}}}, sr, &recalcRepoFake{}, newAnnualRepoFake(), &auditRepoFake{}, &fakeClock{t: now})
	_, err := app.Recalculate(context.Background(), RecalcInput{TenantID: "t1", CycleID: "c1", RuleVersionID: "rv", TriggerReason: "error"})
	if err == nil {
		t.Fatal("summary repository error was swallowed")
	}
}

type errorSummaryRepo struct{}

func (*errorSummaryRepo) GetLatest(context.Context, string, string) (domain.Summary, error) {
	return domain.Summary{}, domain.NewErr(domain.CodeAborted, "storage down")
}
func (*errorSummaryRepo) Save(context.Context, domain.Summary) (domain.Summary, error) {
	return domain.Summary{}, nil
}
func (*errorSummaryRepo) List(context.Context, string, string, int) ([]domain.Summary, error) {
	return nil, nil
}
