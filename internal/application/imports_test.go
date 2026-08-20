package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// fakeClock is a controllable Clock used in tests.
type fakeClock struct {
	t time.Time
}

func (f *fakeClock) Now() time.Time { return f.t }

func newFakeClock() *fakeClock { return &fakeClock{t: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)} }

// makeProject creates a valid Project for tests.
func makeProject(t *testing.T, code string) domain.Project {
	t.Helper()
	return domain.Project{
		ID:           domain.NewID(),
		TenantID:     "tenant-1",
		Code:         code,
		Name:         "项目 " + code,
		Sponsor:      "示例资助方",
		AnnualBudget: 10_000_000,
		StartYear:    2026,
		EndYear:      2027,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestImportsApp_DedupAndTrace(t *testing.T) {
	clk := newFakeClock()
	entries := &entryRepoFake{}
	audit := &auditRepoFake{}
	app := NewImportsApp(entries, audit, clk)

	now := clk.Now()
	cycleID := "cycle-1"
	batchID := "batch-1"
	projectID := "proj-1"

	row := ImportEntryInput{
		TenantID: "tenant-1", CycleID: cycleID, BatchID: batchID, ProjectID: projectID,
		SourceID: "S-001", Source: domain.EntrySourceImport,
		PayeePartyID: "py1", PayerPartyID: "pp1",
		Amount: 100_00, Currency: "CNY", OccurredAt: now,
		Metadata: map[string]string{"note": "trace-A"},
	}

	// First import — creates the entry.
	sum1, _, err := app.ImportEntries(context.Background(), "operator-1", []ImportEntryInput{row})
	if err != nil {
		t.Fatalf("first import: %v", err)
	}
	if sum1.Created != 1 {
		t.Fatalf("expected 1 created, got %d", sum1.Created)
	}

	// Re-import with the same business key — must update (dedup).
	row.Metadata = map[string]string{"note": "trace-B"}
	sum2, _, err := app.ImportEntries(context.Background(), "operator-1", []ImportEntryInput{row})
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if sum2.Updated != 1 || sum2.Created != 0 {
		t.Fatalf("expected 1 updated 0 created, got created=%d updated=%d", sum2.Created, sum2.Updated)
	}

	// Source trace: audit log must have an entry for the import.
	if len(audit.entries) == 0 {
		t.Fatal("audit log must contain the import entry")
	}
	// Two imports → two audit entries; both must carry the import action.
	if len(audit.entries) != 2 {
		t.Fatalf("expected 2 audit entries, got %d", len(audit.entries))
	}
	for i, a := range audit.entries {
		if a.Action != domain.AuditActionImport {
			t.Errorf("audit[%d]: expected action import, got %s", i, a.Action)
		}
	}
	// The first audit entry reports created=1; the second reports updated=1.
	if audit.entries[0].Detail["created"] != "1" || audit.entries[0].Detail["updated"] != "0" {
		t.Errorf("audit[0] detail: %+v", audit.entries[0].Detail)
	}
	if audit.entries[1].Detail["created"] != "0" || audit.entries[1].Detail["updated"] != "1" {
		t.Errorf("audit[1] detail: %+v", audit.entries[1].Detail)
	}

	// Different source_id produces a new fingerprint and is created.
	row2 := row
	row2.SourceID = "S-002"
	sum3, _, err := app.ImportEntries(context.Background(), "operator-1", []ImportEntryInput{row2})
	if err != nil {
		t.Fatalf("third import: %v", err)
	}
	if sum3.Created != 1 {
		t.Fatalf("expected 1 created for new source_id, got %d", sum3.Created)
	}
}

func TestImportsApp_IntraBatchDedup(t *testing.T) {
	clk := newFakeClock()
	entries := &entryRepoFake{}
	audit := &auditRepoFake{}
	app := NewImportsApp(entries, audit, clk)
	now := clk.Now()

	// Two rows with the same business key — intra-batch dedup.
	rows := []ImportEntryInput{
		{
			TenantID: "t1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
			SourceID: "S-x", Source: domain.EntrySourceImport,
			PayeePartyID: "py", PayerPartyID: "pp",
			Amount: 100, Currency: "CNY", OccurredAt: now,
		},
		{
			TenantID: "t1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
			SourceID: "S-x", Source: domain.EntrySourceImport,
			PayeePartyID: "py", PayerPartyID: "pp",
			Amount: 100, Currency: "CNY", OccurredAt: now,
		},
	}
	sum, _, err := app.ImportEntries(context.Background(), "actor", rows)
	if err != nil {
		t.Fatalf("intra-batch dedup: %v", err)
	}
	if sum.Created != 1 {
		t.Fatalf("expected 1 created for intra-batch duplicate, got %d", sum.Created)
	}
}

func TestImportsApp_ValidationErrors(t *testing.T) {
	clk := newFakeClock()
	entries := &entryRepoFake{}
	audit := &auditRepoFake{}
	app := NewImportsApp(entries, audit, clk)

	t.Run("empty rows", func(t *testing.T) {
		_, _, err := app.ImportEntries(context.Background(), "actor", nil)
		if err == nil || !strings.Contains(err.Error(), "no rows to import") {
			t.Fatalf("expected empty rows error, got %v", err)
		}
	})

	t.Run("mismatched tenant", func(t *testing.T) {
		_, _, err := app.ImportEntries(context.Background(), "actor", []ImportEntryInput{
			{TenantID: "t1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
				SourceID: "s", Source: domain.EntrySourceImport,
				PayeePartyID: "py", PayerPartyID: "pp",
				Amount: 1, Currency: "CNY", OccurredAt: clk.Now()},
			{TenantID: "t2", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
				SourceID: "s", Source: domain.EntrySourceImport,
				PayeePartyID: "py", PayerPartyID: "pp",
				Amount: 1, Currency: "CNY", OccurredAt: clk.Now()},
		})
		if err == nil || !strings.Contains(err.Error(), "mismatched tenant") {
			t.Fatalf("expected mismatched tenant error, got %v", err)
		}
	})

	t.Run("invalid amount negative", func(t *testing.T) {
		_, _, err := app.ImportEntries(context.Background(), "actor", []ImportEntryInput{
			{TenantID: "t1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
				SourceID: "s", Source: domain.EntrySourceImport,
				PayeePartyID: "py", PayerPartyID: "pp",
				Amount: -1, Currency: "CNY", OccurredAt: clk.Now()},
		})
		if err == nil {
			t.Fatalf("expected validation error for negative amount")
		}
	})
}

// === Fakes ===

type entryRepoFake struct {
	mu         sync.Mutex
	entries    []domain.SettlementEntry
	upsertErr  error
	listErr    error
	listByCyc  error
	createErr  error
	getErr     error
	upsertCnt  int64
	listCalled atomic.Int32
}

func (e *entryRepoFake) UpsertBatch(ctx context.Context, entries []domain.SettlementEntry) (UpsertSummary, error) {
	if e.upsertErr != nil {
		return UpsertSummary{}, e.upsertErr
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.upsertCnt++
	var s UpsertSummary
	for _, ent := range entries {
		found := false
		for i, ex := range e.entries {
			if ex.SourceFingerprint == ent.SourceFingerprint {
				ent.ID = ex.ID
				ent.CreatedAt = ex.CreatedAt
				e.entries[i] = ent
				s.Updated++
				found = true
				break
			}
		}
		if !found {
			e.entries = append(e.entries, ent)
			s.Created++
		}
	}
	return s, nil
}

func (e *entryRepoFake) Get(ctx context.Context, tenantID, id string) (domain.SettlementEntry, error) {
	if e.getErr != nil {
		return domain.SettlementEntry{}, e.getErr
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, ent := range e.entries {
		if ent.ID == id && ent.TenantID == tenantID {
			return ent, nil
		}
	}
	return domain.SettlementEntry{}, domain.NewErr(domain.CodeNotFound, "not found")
}

func (e *entryRepoFake) List(ctx context.Context, q EntryListQuery) ([]domain.SettlementEntry, PageResult, error) {
	e.listCalled.Add(1)
	if e.listErr != nil {
		return nil, PageResult{}, e.listErr
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]domain.SettlementEntry, 0)
	for _, ent := range e.entries {
		if ent.TenantID != q.TenantID {
			continue
		}
		if q.CycleID != "" && ent.CycleID != q.CycleID {
			continue
		}
		out = append(out, ent)
	}
	return out, PageResult{Page: 1, PageSize: q.PageSize, Total: len(out), HasNext: false}, nil
}

func (e *entryRepoFake) ListByCycle(ctx context.Context, tenantID, cycleID string) ([]domain.SettlementEntry, error) {
	if e.listByCyc != nil {
		return nil, e.listByCyc
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]domain.SettlementEntry, 0)
	for _, ent := range e.entries {
		if ent.TenantID == tenantID && ent.CycleID == cycleID {
			out = append(out, ent)
		}
	}
	return out, nil
}

type auditRepoFake struct {
	entries []domain.AuditEntry
	err     error
}

func (a *auditRepoFake) Append(ctx context.Context, e domain.AuditEntry) (domain.AuditEntry, error) {
	if a.err != nil {
		return domain.AuditEntry{}, a.err
	}
	a.entries = append(a.entries, e)
	return e, nil
}

func (a *auditRepoFake) List(ctx context.Context, q ListQuery) ([]domain.AuditEntry, PageResult, error) {
	out := make([]domain.AuditEntry, 0)
	for _, e := range a.entries {
		if e.TenantID != q.TenantID {
			continue
		}
		out = append(out, e)
	}
	return out, PageResult{Page: 1, PageSize: q.PageSize, Total: len(out), HasNext: false}, nil
}

// ruleVersionRepoFake is a minimal rule version repo for evaluate tests.
type ruleVersionRepoFake struct {
	rv  domain.RuleVersion
	err error
}

func (r *ruleVersionRepoFake) Create(ctx context.Context, rv domain.RuleVersion) (domain.RuleVersion, error) {
	return rv, nil
}
func (r *ruleVersionRepoFake) Get(ctx context.Context, tenantID, id string) (domain.RuleVersion, error) {
	return r.rv, r.err
}
func (r *ruleVersionRepoFake) GetByCode(ctx context.Context, tenantID, code string) (domain.RuleVersion, error) {
	return r.rv, r.err
}
func (r *ruleVersionRepoFake) List(ctx context.Context, q ListQuery) ([]domain.RuleVersion, PageResult, error) {
	return []domain.RuleVersion{r.rv}, PageResult{Page: 1, PageSize: 20, Total: 1, HasNext: false}, nil
}
func (r *ruleVersionRepoFake) Update(ctx context.Context, rv domain.RuleVersion) (domain.RuleVersion, error) {
	return rv, nil
}

// exceptionRepoFake is a minimal exception repo for evaluate / lifecycle tests.
type exceptionRepoFake struct {
	mu        sync.Mutex
	items     []domain.Exception
	createErr error
	updateErr error
	staleVer  bool
}

func (e *exceptionRepoFake) Create(ctx context.Context, ex domain.Exception) (domain.Exception, error) {
	if e.createErr != nil {
		return domain.Exception{}, e.createErr
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.items = append(e.items, ex)
	return ex, nil
}

func (e *exceptionRepoFake) Get(ctx context.Context, tenantID, id string) (domain.Exception, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, ex := range e.items {
		if ex.ID == id && ex.TenantID == tenantID {
			return ex, nil
		}
	}
	return domain.Exception{}, domain.NewErr(domain.CodeNotFound, "not found")
}

func (e *exceptionRepoFake) Update(ctx context.Context, ex domain.Exception) (domain.Exception, error) {
	if e.updateErr != nil {
		return domain.Exception{}, e.updateErr
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, exi := range e.items {
		if exi.ID == ex.ID && exi.TenantID == ex.TenantID {
			if e.staleVer && ex.Version != exi.Version+1 {
				return domain.Exception{}, domain.NewErr(domain.CodeAborted, "stale")
			}
			e.items[i] = ex
			return ex, nil
		}
	}
	return domain.Exception{}, domain.NewErr(domain.CodeNotFound, "not found")
}

func (e *exceptionRepoFake) List(ctx context.Context, q ExceptionListQuery) ([]domain.Exception, PageResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]domain.Exception, 0)
	for _, ex := range e.items {
		if ex.TenantID != q.TenantID {
			continue
		}
		if q.CycleID != "" && ex.CycleID != q.CycleID {
			continue
		}
		if q.Status != "" && string(ex.Status) != q.Status {
			continue
		}
		out = append(out, ex)
	}
	return out, PageResult{Page: 1, PageSize: q.PageSize, Total: len(out), HasNext: false}, nil
}

func (e *exceptionRepoFake) ListByAssignee(ctx context.Context, tenantID, assigneeID string) ([]domain.Exception, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]domain.Exception, 0)
	for _, ex := range e.items {
		if ex.TenantID == tenantID && ex.AssigneeID == assigneeID {
			out = append(out, ex)
		}
	}
	return out, nil
}

func (e *exceptionRepoFake) ListByCycle(ctx context.Context, tenantID, cycleID string) ([]domain.Exception, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]domain.Exception, 0)
	for _, ex := range e.items {
		if ex.TenantID == tenantID && ex.CycleID == cycleID {
			out = append(out, ex)
		}
	}
	return out, nil
}

// Helper to silence unused-import in errors (kept for future assertion helpers).
var _ = errors.New
