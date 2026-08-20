package application

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// === Summary/Recalc fakes ===

type cycleRepoFake struct {
	cycle domain.SettlementCycle
	err   error
}

func (c *cycleRepoFake) Create(ctx context.Context, cy domain.SettlementCycle) (domain.SettlementCycle, error) {
	return cy, nil
}
func (c *cycleRepoFake) Get(ctx context.Context, tenantID, id string) (domain.SettlementCycle, error) {
	return c.cycle, c.err
}
func (c *cycleRepoFake) List(ctx context.Context, q ListQuery) ([]domain.SettlementCycle, PageResult, error) {
	return []domain.SettlementCycle{c.cycle}, PageResult{Page: 1, PageSize: 20, Total: 1}, nil
}
func (c *cycleRepoFake) Update(ctx context.Context, cy domain.SettlementCycle) (domain.SettlementCycle, error) {
	return cy, nil
}

type summaryRepoFake struct {
	items []domain.Summary
}

func (s *summaryRepoFake) GetLatest(ctx context.Context, tenantID, cycleID string) (domain.Summary, error) {
	var latest domain.Summary
	found := false
	for _, sm := range s.items {
		if sm.TenantID != tenantID || sm.CycleID != cycleID {
			continue
		}
		if !found || sm.Version > latest.Version {
			latest = sm
			found = true
		}
	}
	if !found {
		return domain.Summary{}, domain.NewErr(domain.CodeNotFound, "no summary")
	}
	return latest, nil
}

func (s *summaryRepoFake) Save(ctx context.Context, sm domain.Summary) (domain.Summary, error) {
	s.items = append(s.items, sm)
	return sm, nil
}

func (s *summaryRepoFake) List(ctx context.Context, tenantID, cycleID string, limit int) ([]domain.Summary, error) {
	out := make([]domain.Summary, 0)
	for _, sm := range s.items {
		if sm.TenantID == tenantID && sm.CycleID == cycleID {
			out = append(out, sm)
		}
	}
	return out, nil
}

type recalcRepoFake struct {
	items []domain.RecalculationBatch
}

func (r *recalcRepoFake) Create(ctx context.Context, rb domain.RecalculationBatch) (domain.RecalculationBatch, error) {
	r.items = append(r.items, rb)
	return rb, nil
}
func (r *recalcRepoFake) Update(ctx context.Context, rb domain.RecalculationBatch) (domain.RecalculationBatch, error) {
	for i, x := range r.items {
		if x.ID == rb.ID {
			r.items[i] = rb
			return rb, nil
		}
	}
	r.items = append(r.items, rb)
	return rb, nil
}
func (r *recalcRepoFake) Get(ctx context.Context, tenantID, id string) (domain.RecalculationBatch, error) {
	for _, x := range r.items {
		if x.ID == id {
			return x, nil
		}
	}
	return domain.RecalculationBatch{}, domain.NewErr(domain.CodeNotFound, "not found")
}
func (r *recalcRepoFake) List(ctx context.Context, q ListQuery) ([]domain.RecalculationBatch, PageResult, error) {
	return r.items, PageResult{Page: 1, PageSize: 20, Total: len(r.items)}, nil
}

type annualRepoFake struct {
	mu      sync.Mutex
	byKey   map[string]domain.AnnualAccumulator
	adjList []domain.Adjustment
}

func newAnnualRepoFake() *annualRepoFake {
	return &annualRepoFake{byKey: map[string]domain.AnnualAccumulator{}}
}

func (a *annualRepoFake) Get(ctx context.Context, tenantID, projectID string, year int) (domain.AnnualAccumulator, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	k := projectID + "|" + itoa(year)
	acc, ok := a.byKey[k]
	if !ok {
		return domain.AnnualAccumulator{}, domain.NewErr(domain.CodeNotFound, "not found")
	}
	return acc, nil
}

func (a *annualRepoFake) ApplyAdjustment(ctx context.Context, adj domain.Adjustment) (domain.AnnualAccumulator, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	k := adj.ProjectID + "|" + itoa(adj.Year)
	acc := a.byKey[k]
	acc.ProjectID = adj.ProjectID
	acc.Year = adj.Year
	acc = acc.ApplyAdjustment(adj)
	a.byKey[k] = acc
	a.adjList = append(a.adjList, adj)
	return acc, nil
}

func (a *annualRepoFake) ListAdjustments(ctx context.Context, tenantID, projectID string, year int) ([]domain.Adjustment, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]domain.Adjustment, 0)
	for _, x := range a.adjList {
		if x.ProjectID == projectID && x.Year == year {
			out = append(out, x)
		}
	}
	return out, nil
}

func TestSummaryApp_UnresolvedExcludedFromEligible(t *testing.T) {
	clk := newFakeClock()
	now := clk.Now()

	cycle := domain.SettlementCycle{ID: "c1", TenantID: "t1", ProjectID: "p1", FundingBatchID: "b1", Year: 2026, Quarter: 1, StartDate: now, EndDate: now}
	rv := domain.RuleVersion{ID: "rv1", TenantID: "t1", Code: "RV-1", ProjectID: "p1", Status: domain.RuleVersionStatusPublished,
		Rules: []domain.RuleDefinition{{ID: "r1", Code: "AMOUNT_ZERO", Severity: domain.SeverityHigh, Expression: "amount == 0"}}}

	// Two entries: e1 with unresolved exception, e2 with resolved exception.
	entries := []domain.SettlementEntry{
		{ID: "e1", TenantID: "t1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
			Source: domain.EntrySourceImport, PayeePartyID: "py", PayerPartyID: "pp",
			Amount: 1000, Currency: "CNY", OccurredAt: now, SourceFingerprint: "fp1"},
		{ID: "e2", TenantID: "t1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
			Source: domain.EntrySourceImport, PayeePartyID: "py", PayerPartyID: "pp",
			Amount: 2000, Currency: "CNY", OccurredAt: now, SourceFingerprint: "fp2"},
	}

	excs := []domain.Exception{
		{ID: "ex1", TenantID: "t1", CycleID: "c1", EntryID: "e1", RuleVersionID: "rv1",
			RuleCode: "AMOUNT_ZERO", Severity: domain.SeverityHigh, Title: "x", Status: domain.ExceptionStatusPending,
			Snapshot: domain.ExceptionSnapshot{SnapshotAt: now}},
		{ID: "ex2", TenantID: "t1", CycleID: "c1", EntryID: "e2", RuleVersionID: "rv1",
			RuleCode: "AMOUNT_ZERO", Severity: domain.SeverityHigh, Title: "x", Status: domain.ExceptionStatusResolved,
			Snapshot: domain.ExceptionSnapshot{SnapshotAt: now}, ResolvedAt: now},
	}

	app := NewSummaryApp(
		&cycleRepoFake{cycle: cycle},
		&entryRepoFake{entries: entries},
		&exceptionRepoFake{items: excs},
		&ruleVersionRepoFake{rv: rv},
		&summaryRepoFake{},
		&recalcRepoFake{},
		newAnnualRepoFake(),
		&auditRepoFake{},
		clk,
	)

	res, err := app.Recalculate(context.Background(), RecalcInput{
		TenantID: "t1", CycleID: "c1", RuleVersionID: "rv1",
		ActorID: "admin", TriggerReason: "test",
	})
	if err != nil {
		t.Fatalf("recalc: %v", err)
	}
	// Total = 1000 + 2000 = 3000; approved = 2000 (e1 excluded due to pending exception).
	if res.Summary.TotalAmountCents != 3000 {
		t.Errorf("expected total 3000, got %d", res.Summary.TotalAmountCents)
	}
	if res.Summary.ApprovedAmountCents != 2000 {
		t.Errorf("expected approved 2000, got %d", res.Summary.ApprovedAmountCents)
	}
	if res.Summary.PendingAmountCents != 1000 {
		t.Errorf("expected pending 1000, got %d", res.Summary.PendingAmountCents)
	}
}

func TestSummaryApp_SnapshotPreservingRecalc(t *testing.T) {
	clk := newFakeClock()
	now := clk.Now()
	cycle := domain.SettlementCycle{ID: "c1", TenantID: "t1", ProjectID: "p1", FundingBatchID: "b1", Year: 2026, Quarter: 1, StartDate: now, EndDate: now}
	rv := domain.RuleVersion{ID: "rv1", TenantID: "t1", Code: "RV-1", ProjectID: "p1", Status: domain.RuleVersionStatusPublished,
		Rules: []domain.RuleDefinition{{ID: "r1", Code: "AMOUNT_ZERO", Severity: domain.SeverityHigh, Expression: "amount == 0"}}}

	entries := []domain.SettlementEntry{
		{ID: "e1", TenantID: "t1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
			Source: domain.EntrySourceImport, PayeePartyID: "py", PayerPartyID: "pp",
			Amount: 5000, Currency: "CNY", OccurredAt: now, SourceFingerprint: "fp1"},
	}

	// First recalc: no exceptions → approved = 5000.
	app := NewSummaryApp(
		&cycleRepoFake{cycle: cycle},
		&entryRepoFake{entries: entries},
		&exceptionRepoFake{items: nil},
		&ruleVersionRepoFake{rv: rv},
		&summaryRepoFake{},
		&recalcRepoFake{},
		newAnnualRepoFake(),
		&auditRepoFake{},
		clk,
	)
	res1, err := app.Recalculate(context.Background(), RecalcInput{
		TenantID: "t1", CycleID: "c1", RuleVersionID: "rv1",
		ActorID: "admin", TriggerReason: "first",
	})
	if err != nil {
		t.Fatalf("first recalc: %v", err)
	}
	if res1.Summary.ApprovedAmountCents != 5000 {
		t.Fatalf("expected 5000 approved, got %d", res1.Summary.ApprovedAmountCents)
	}

	// Second recalc after adding an unresolved exception — approved must drop.
	app2 := NewSummaryApp(
		&cycleRepoFake{cycle: cycle},
		&entryRepoFake{entries: entries},
		&exceptionRepoFake{items: []domain.Exception{
			{ID: "ex1", TenantID: "t1", CycleID: "c1", EntryID: "e1", RuleVersionID: "rv1",
				RuleCode: "AMOUNT_ZERO", Severity: domain.SeverityHigh, Title: "x", Status: domain.ExceptionStatusPending,
				Snapshot: domain.ExceptionSnapshot{SnapshotAt: now}},
		}},
		&ruleVersionRepoFake{rv: rv},
		&summaryRepoFake{items: []domain.Summary{res1.Summary}}, // preserve prior snapshot
		&recalcRepoFake{},
		newAnnualRepoFake(),
		&auditRepoFake{},
		clk,
	)
	res2, err := app2.Recalculate(context.Background(), RecalcInput{
		TenantID: "t1", CycleID: "c1", RuleVersionID: "rv1",
		ActorID: "admin", TriggerReason: "second — exception added",
	})
	if err != nil {
		t.Fatalf("second recalc: %v", err)
	}
	if res2.Summary.ApprovedAmountCents != 0 {
		t.Errorf("expected approved to drop to 0, got %d", res2.Summary.ApprovedAmountCents)
	}
	if res2.Summary.Version != res1.Summary.Version+1 {
		t.Errorf("expected version %d, got %d", res1.Summary.Version+1, res2.Summary.Version)
	}
	if res2.Summary.DiffBasis.PreviousApprovedCents != 5000 {
		t.Errorf("expected previous approved 5000, got %d", res2.Summary.DiffBasis.PreviousApprovedCents)
	}
	if res2.Summary.DiffBasis.DeltaApprovedCents != -5000 {
		t.Errorf("expected delta -5000, got %d", res2.Summary.DiffBasis.DeltaApprovedCents)
	}
}

func TestSummaryApp_AnnualAdjustmentAndOverrun(t *testing.T) {
	clk := newFakeClock()
	annual := newAnnualRepoFake()
	app := NewSummaryApp(
		&cycleRepoFake{},
		&entryRepoFake{},
		&exceptionRepoFake{},
		&ruleVersionRepoFake{},
		&summaryRepoFake{},
		&recalcRepoFake{},
		annual,
		&auditRepoFake{},
		clk,
	)

	// Two adjustments: +3000 then +8000 → disbursed 11000, no budget set yet → overrun 11000.
	_, err := app.ApplyAdjustment(context.Background(), AdjustAnnualInput{
		TenantID: "t1", ProjectID: "p1", Year: 2026, DeltaCents: 3000, Reason: "首批", ActorID: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	acc, err := app.ApplyAdjustment(context.Background(), AdjustAnnualInput{
		TenantID: "t1", ProjectID: "p1", Year: 2026, DeltaCents: 8000, Reason: "追加", ActorID: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if acc.DisbursedCents != 11000 {
		t.Errorf("expected 11000 disbursed, got %d", acc.DisbursedCents)
	}
	if acc.OverrunCents() != 11000 {
		t.Errorf("expected 11000 overrun, got %d", acc.OverrunCents())
	}

	// Reason is mandatory.
	_, err = app.ApplyAdjustment(context.Background(), AdjustAnnualInput{
		TenantID: "t1", ProjectID: "p1", Year: 2026, DeltaCents: -1000,
	})
	if err == nil || !strings.Contains(err.Error(), "reason required") {
		t.Fatalf("expected reason required, got %v", err)
	}

	// List adjustments records both.
	adjs, err := app.ListAdjustments(context.Background(), "t1", "p1", 2026)
	if err != nil {
		t.Fatal(err)
	}
	if len(adjs) != 2 {
		t.Errorf("expected 2 adjustments, got %d", len(adjs))
	}
}

func TestSummaryApp_RecalcInputValidation(t *testing.T) {
	app := NewSummaryApp(&cycleRepoFake{}, &entryRepoFake{}, &exceptionRepoFake{}, &ruleVersionRepoFake{}, &summaryRepoFake{}, &recalcRepoFake{}, newAnnualRepoFake(), &auditRepoFake{}, newFakeClock())

	_, err := app.Recalculate(context.Background(), RecalcInput{})
	if err == nil || !strings.Contains(err.Error(), "tenant id required") {
		t.Fatalf("expected tenant id error, got %v", err)
	}
}
