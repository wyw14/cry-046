package application

import (
	"context"
	"fmt"
	"testing"

	"github.com/welfare/settlement-resolver/internal/domain"
)

func TestEvaluateCycle_RuleVersionIsPartOfIdempotencyKey(t *testing.T) {
	clk := newFakeClock()
	now := clk.Now()
	makeRule := func(id string) domain.RuleVersion {
		return domain.RuleVersion{ID: id, TenantID: "tenant-a", Code: "AMOUNT_ZERO_V1", ProjectID: "project-a", Description: "published settlement rule", Rules: []domain.RuleDefinition{{ID: "amount-zero", Code: "AMOUNT_ZERO", Severity: domain.SeverityHigh, Expression: "amount == 0", DeadlineHours: 24}}, Status: domain.RuleVersionStatusPublished, PublishedAt: now, Version: 1}
	}
	rules := &versionedRuleRepo{items: map[string]domain.RuleVersion{"rv-old": makeRule("rv-old"), "rv-new": makeRule("rv-new")}}
	entries := &entryRepoFake{entries: []domain.SettlementEntry{{ID: "entry-1", TenantID: "tenant-a", CycleID: "cycle-a", BatchID: "batch-a", ProjectID: "project-a", Source: domain.EntrySourceImport, PayeePartyID: "payee", PayerPartyID: "payer", Amount: 0, Currency: "CNY", OccurredAt: now, SourceFingerprint: "fp-entry-1"}}}
	exceptions := &exceptionRepoFake{}
	app := NewEvaluateApp(rules, entries, exceptions, &auditRepoFake{}, clk)
	first, err := app.EvaluateCycle(context.Background(), EvaluateCycleInput{TenantID: "tenant-a", CycleID: "cycle-a", RuleVersionID: "rv-old", ActorID: "operator-a"})
	if err != nil || first.CreatedExceptions != 1 { t.Fatalf("first evaluation: created=%d err=%v", first.CreatedExceptions, err) }
	replay, err := app.EvaluateCycle(context.Background(), EvaluateCycleInput{TenantID: "tenant-a", CycleID: "cycle-a", RuleVersionID: "rv-old", ActorID: "operator-a"})
	if err != nil || replay.CreatedExceptions != 0 { t.Fatalf("same-version replay: created=%d err=%v", replay.CreatedExceptions, err) }
	second, err := app.EvaluateCycle(context.Background(), EvaluateCycleInput{TenantID: "tenant-a", CycleID: "cycle-a", RuleVersionID: "rv-new", ActorID: "operator-a"})
	if err != nil { t.Fatalf("new-version evaluation: %v", err) }
	if second.CreatedExceptions != 1 || len(exceptions.items) != 2 { t.Fatalf("new rule version was treated as replay: created=%d total=%d", second.CreatedExceptions, len(exceptions.items)) }
	if exceptions.items[0].RuleVersionID != "rv-old" || exceptions.items[1].RuleVersionID != "rv-new" { t.Fatalf("rule version identity not preserved: %#v", exceptions.items) }
}

type versionedRuleRepo struct{ items map[string]domain.RuleVersion }
func (r *versionedRuleRepo) Create(context.Context, domain.RuleVersion) (domain.RuleVersion, error) { return domain.RuleVersion{}, fmt.Errorf("not used") }
func (r *versionedRuleRepo) Get(_ context.Context, tenantID, id string) (domain.RuleVersion, error) { rv, ok := r.items[id]; if !ok || rv.TenantID != tenantID { return domain.RuleVersion{}, fmt.Errorf("rule %s not found", id) }; return rv, nil }
func (r *versionedRuleRepo) GetByCode(context.Context, string, string) (domain.RuleVersion, error) { return domain.RuleVersion{}, fmt.Errorf("not used") }
func (r *versionedRuleRepo) List(context.Context, ListQuery) ([]domain.RuleVersion, PageResult, error) { return nil, PageResult{}, fmt.Errorf("not used") }
func (r *versionedRuleRepo) Update(context.Context, domain.RuleVersion) (domain.RuleVersion, error) { return domain.RuleVersion{}, fmt.Errorf("not used") }
