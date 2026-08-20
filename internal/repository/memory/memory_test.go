package memory

import (
	"context"
	"testing"
	"time"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

func testClock() fixedClock {
	return fixedClock{t: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)}
}

func TestStoreProjects(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	p, err := repos.Projects.Create(ctx, domain.Project{
		ID: "p1", TenantID: "t1", Code: "WS-01", Name: "P1", Sponsor: "S",
		AnnualBudget: 1000, StartYear: 2026, EndYear: 2027,
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.ID != "p1" {
		t.Errorf("expected p1, got %s", p.ID)
	}
	got, err := repos.Projects.Get(ctx, "t1", "p1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Code != "WS-01" {
		t.Errorf("expected WS-01, got %s", got.Code)
	}
	// Duplicate code should fail.
	_, err = repos.Projects.Create(ctx, domain.Project{
		ID: "p2", TenantID: "t1", Code: "WS-01", Name: "P2", Sponsor: "S",
		AnnualBudget: 1000, StartYear: 2026, EndYear: 2027,
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	})
	if err == nil || !domain.IsAlreadyExists(err) {
		t.Fatalf("expected already exists, got %v", err)
	}
	// Update should change the project.
	got.Name = "P1-updated"
	updated, err := repos.Projects.Update(ctx, got)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "P1-updated" {
		t.Errorf("expected P1-updated, got %s", updated.Name)
	}
}

func TestStoreProjectsList(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	for i := 0; i < 5; i++ {
		_, _ = repos.Projects.Create(ctx, domain.Project{
			ID: "p" + string(rune('1'+i)), TenantID: "t1",
			Code: "WS-" + string(rune('0'+i+1)), Name: "P", Sponsor: "S",
			AnnualBudget: 1000, StartYear: 2026, EndYear: 2027,
			CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
		})
	}
	list, page, err := repos.Projects.List(ctx, application.ListQuery{TenantID: "t1", Page: 1, PageSize: 3})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 items, got %d", len(list))
	}
	if page.Total != 5 {
		t.Errorf("expected total 5, got %d", page.Total)
	}
	if !page.HasNext {
		t.Error("expected HasNext")
	}
}

func TestStoreParties(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	p, err := repos.Parties.Create(ctx, domain.Party{
		ID: "py1", TenantID: "t1", Name: "Donor", Type: domain.PartyTypeDonor,
		Contact: "donor@example.com", CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.ID != "py1" {
		t.Errorf("expected py1, got %s", p.ID)
	}
	got, err := repos.Parties.Get(ctx, "t1", "py1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Donor" {
		t.Errorf("expected Donor, got %s", got.Name)
	}
}

func TestStoreBatches(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	b, err := repos.Batches.Create(ctx, domain.FundingBatch{
		ID: "b1", TenantID: "t1", ProjectID: "p1", Code: "FB-01",
		DonorPartyID: "d1", ImplementerPartyID: "i1",
		TotalAmount: 1000_00, Currency: "CNY", DisbursedAt: clk.Now(),
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if b.Code != "FB-01" {
		t.Errorf("expected FB-01, got %s", b.Code)
	}
}

func TestStoreCycles(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	c, err := repos.Cycles.Create(ctx, domain.SettlementCycle{
		ID: "c1", TenantID: "t1", ProjectID: "p1", FundingBatchID: "b1",
		Year: 2026, Quarter: 1, StartDate: clk.Now(), EndDate: clk.Now().AddDate(0, 3, 0),
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if c.ID != "c1" {
		t.Errorf("expected c1, got %s", c.ID)
	}
	closed, err := repos.Cycles.Update(ctx, domain.SettlementCycle{
		ID: "c1", TenantID: "t1", ProjectID: "p1", FundingBatchID: "b1",
		Year: 2026, Quarter: 1, StartDate: clk.Now(), EndDate: clk.Now().AddDate(0, 3, 0),
		ClosedAt: clk.Now(), CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !closed.IsClosed() {
		t.Error("expected closed cycle")
	}
}

func TestStoreRuleVersions(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	rv, err := repos.Rules.Create(ctx, domain.RuleVersion{
		ID: "rv1", TenantID: "t1", Code: "RV-01", ProjectID: "p1",
		Rules:     []domain.RuleDefinition{{Code: "R1", Severity: domain.SeverityHigh, Expression: "amount == 0"}},
		Status:    domain.RuleVersionStatusDraft,
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(), Version: 1,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if rv.Code != "RV-01" {
		t.Errorf("expected RV-01, got %s", rv.Code)
	}
	byCode, err := repos.Rules.GetByCode(ctx, "t1", "RV-01")
	if err != nil {
		t.Fatalf("get by code: %v", err)
	}
	if byCode.ID != "rv1" {
		t.Errorf("expected rv1, got %s", byCode.ID)
	}
	// Update with stale version should fail.
	_, err = repos.Rules.Update(ctx, domain.RuleVersion{
		ID: "rv1", TenantID: "t1", Code: "RV-01", ProjectID: "p1",
		Rules:     []domain.RuleDefinition{{Code: "R1", Severity: domain.SeverityHigh, Expression: "amount == 0"}},
		Status:    domain.RuleVersionStatusPublished,
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(), Version: 5,
	})
	if err == nil || !domain.IsAborted(err) {
		t.Fatalf("expected aborted, got %v", err)
	}
}

func TestStoreEntriesUpsert(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	fp := domain.EntryDedupFingerprint("c1", "b1", "s1", "pp1", "py1", 100, clk.Now())
	e1 := domain.SettlementEntry{
		ID: "e1", TenantID: "t1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
		SourceID: "s1", Source: domain.EntrySourceImport,
		PayeePartyID: "py1", PayerPartyID: "pp1",
		Amount: 100, Currency: "CNY", OccurredAt: clk.Now(),
		SourceFingerprint: fp, CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	}
	summary, err := repos.Entries.UpsertBatch(ctx, []domain.SettlementEntry{e1})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if summary.Created != 1 {
		t.Errorf("expected 1 created, got %d", summary.Created)
	}
	// Upsert same fingerprint: should update, not create.
	e2 := e1
	e2.Amount = 200
	e2.UpdatedAt = clk.Now()
	summary, err = repos.Entries.UpsertBatch(ctx, []domain.SettlementEntry{e2})
	if err != nil {
		t.Fatalf("upsert 2: %v", err)
	}
	if summary.Updated != 1 {
		t.Errorf("expected 1 updated, got %d", summary.Updated)
	}
	got, err := repos.Entries.Get(ctx, "t1", "e1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Amount != 200 {
		t.Errorf("expected 200, got %d", got.Amount)
	}
}

func TestStoreEntriesListByCycle(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	for i := 0; i < 3; i++ {
		fp := domain.EntryDedupFingerprint("c1", "b1", "s"+string(rune('0'+i)), "pp1", "py1", 100, clk.Now())
		_, _ = repos.Entries.UpsertBatch(ctx, []domain.SettlementEntry{
			{
				ID: "e" + string(rune('0'+i)), TenantID: "t1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
				SourceID: "s" + string(rune('0'+i)), Source: domain.EntrySourceImport,
				PayeePartyID: "py1", PayerPartyID: "pp1",
				Amount: 100, Currency: "CNY", OccurredAt: clk.Now(),
				SourceFingerprint: fp, CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
			},
		})
	}
	list, err := repos.Entries.ListByCycle(ctx, "t1", "c1")
	if err != nil {
		t.Fatalf("list by cycle: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 entries, got %d", len(list))
	}
}

func TestStoreExceptions(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	ex := domain.Exception{
		ID: "ex1", TenantID: "t1", CycleID: "c1", EntryID: "e1",
		RuleVersionID: "rv1", RuleCode: "R1", Severity: domain.SeverityHigh,
		Title: "T", Status: domain.ExceptionStatusPending,
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(), Version: 1,
		Snapshot: domain.ExceptionSnapshot{SnapshotAt: clk.Now()},
	}
	_, err := repos.Exceptions.Create(ctx, ex)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := repos.Exceptions.Get(ctx, "t1", "ex1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "T" {
		t.Errorf("expected T, got %s", got.Title)
	}
	// Stale version should fail.
	got.Version = 5
	got.Title = "updated"
	_, err = repos.Exceptions.Update(ctx, got)
	if err == nil || !domain.IsAborted(err) {
		t.Fatalf("expected aborted, got %v", err)
	}
}

func TestStoreExceptionsListByAssignee(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	for i := 0; i < 3; i++ {
		ex := domain.Exception{
			ID: "ex" + string(rune('0'+i)), TenantID: "t1", CycleID: "c1", EntryID: "e1",
			RuleVersionID: "rv1", RuleCode: "R1", Severity: domain.SeverityHigh,
			Title: "T", Status: domain.ExceptionStatusProcessing, AssigneeID: "u1",
			CreatedAt: clk.Now(), UpdatedAt: clk.Now(), Version: 1,
			Snapshot: domain.ExceptionSnapshot{SnapshotAt: clk.Now()},
		}
		_, _ = repos.Exceptions.Create(ctx, ex)
	}
	list, err := repos.Exceptions.ListByAssignee(ctx, "t1", "u1")
	if err != nil {
		t.Fatalf("list by assignee: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 exceptions, got %d", len(list))
	}
}

func TestStoreSummaries(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	s1 := domain.Summary{
		ID: "s1", TenantID: "t1", CycleID: "c1", RuleVersionID: "rv1",
		ComputedAt: clk.Now(), TotalEntries: 2, TotalAmountCents: 1000,
		ApprovedAmountCents: 500, Version: 1,
	}
	_, err := repos.Summaries.Save(ctx, s1)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	s2 := s1
	s2.ID = "s2"
	s2.Version = 2
	s2.ApprovedAmountCents = 800
	_, _ = repos.Summaries.Save(ctx, s2)
	latest, err := repos.Summaries.GetLatest(ctx, "t1", "c1")
	if err != nil {
		t.Fatalf("get latest: %v", err)
	}
	if latest.Version != 2 {
		t.Errorf("expected version 2, got %d", latest.Version)
	}
	history, err := repos.Summaries.List(ctx, "t1", "c1", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("expected 2 summaries, got %d", len(history))
	}
}

func TestStoreAnnualRepo(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	adj1 := domain.Adjustment{
		ID: "a1", ProjectID: "p1", Year: 2026, DeltaCents: 3000,
		Reason: "first", TriggeredBy: "admin", CreatedAt: clk.Now(),
	}
	acc, err := repos.Annuals.ApplyAdjustment(ctx, adj1)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if acc.DisbursedCents != 3000 {
		t.Errorf("expected 3000 disbursed, got %d", acc.DisbursedCents)
	}
	got, err := repos.Annuals.Get(ctx, "t1", "p1", 2026)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.DisbursedCents != 3000 {
		t.Errorf("expected 3000, got %d", got.DisbursedCents)
	}
	adjs, err := repos.Annuals.ListAdjustments(ctx, "t1", "p1", 2026)
	if err != nil {
		t.Fatalf("list adjustments: %v", err)
	}
	if len(adjs) != 1 {
		t.Errorf("expected 1 adjustment, got %d", len(adjs))
	}
}

func TestStoreRecalcs(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	rb := domain.RecalculationBatch{
		ID: "rb1", TenantID: "t1", CycleID: "c1", RuleVersionID: "rv1",
		InputDigest: "abc", TriggerReason: "test", StartedAt: clk.Now(),
		Status:    domain.RecalcStatusRunning,
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	}
	_, err := repos.Recalcs.Create(ctx, rb)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	rb.Status = domain.RecalcStatusCompleted
	rb.FinishedAt = clk.Now()
	_, err = repos.Recalcs.Update(ctx, rb)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err := repos.Recalcs.Get(ctx, "t1", "rb1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != domain.RecalcStatusCompleted {
		t.Errorf("expected completed, got %s", got.Status)
	}
}

func TestStoreAudits(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	e := domain.AuditEntry{
		ID: "a1", TenantID: "t1", ActorID: "u1", Action: domain.AuditActionImport,
		EntityType: "settlement_entry", EntityID: "e1", CreatedAt: clk.Now(),
	}
	_, err := repos.Audits.Append(ctx, e)
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	list, _, err := repos.Audits.List(ctx, application.ListQuery{TenantID: "t1", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 audit, got %d", len(list))
	}
}

func TestStoreUsers(t *testing.T) {
	store := NewStore()
	repos := New(store)
	ctx := context.Background()
	clk := testClock()
	u := domain.User{
		ID: "u1", TenantID: "t1", Username: "admin", DisplayName: "Admin",
		Email: "admin@example.com", Role: domain.RoleAdmin,
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	}
	_, err := repos.Users.Create(ctx, u)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := repos.Users.GetByUsername(ctx, "t1", "admin")
	if err != nil {
		t.Fatalf("get by username: %v", err)
	}
	if got.ID != "u1" {
		t.Errorf("expected u1, got %s", got.ID)
	}
	// Duplicate username should fail.
	_, err = repos.Users.Create(ctx, domain.User{
		ID: "u2", TenantID: "t1", Username: "admin", Role: domain.RoleAdmin,
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	})
	if err == nil || !domain.IsAlreadyExists(err) {
		t.Fatalf("expected already exists, got %v", err)
	}
}

func TestStoreUnitDo(t *testing.T) {
	store := NewStore()
	repos := New(store)
	unit := NewUnit(repos)
	ctx := context.Background()
	clk := testClock()
	err := unit.Do(ctx, func(ctx context.Context, uow application.UnitOfWork) error {
		_, err := uow.Projects().Create(ctx, domain.Project{
			ID: "p1", TenantID: "t1", Code: "WS-01", Name: "P1", Sponsor: "S",
			AnnualBudget: 1000, StartYear: 2026, EndYear: 2027,
			CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
		})
		return err
	})
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	got, err := repos.Projects.Get(ctx, "t1", "p1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Code != "WS-01" {
		t.Errorf("expected WS-01, got %s", got.Code)
	}
}
