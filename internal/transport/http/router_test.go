package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
	"github.com/welfare/settlement-resolver/internal/repository/memory"
)

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

func newTestRouter(t *testing.T) (Router, *memory.Repositories, fixedClock) {
	t.Helper()
	store := memory.NewStore()
	repos := memory.New(store)
	clk := fixedClock{t: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)}
	projectsApp := application.NewProjectsApp(repos.Projects, repos.Parties, repos.Batches, repos.Cycles, repos.Rules, clk)
	partiesApp := application.NewPartiesApp(repos.Parties, clk)
	batchesApp := application.NewBatchesApp(repos.Batches, clk)
	cyclesApp := application.NewCyclesApp(repos.Cycles, clk)
	rulesApp := application.NewRulesApp(repos.Rules, clk)
	importsApp := application.NewImportsApp(repos.Entries, repos.Audits, clk)
	evaluateApp := application.NewEvaluateApp(repos.Rules, repos.Entries, repos.Exceptions, repos.Audits, clk)
	exceptionsApp := application.NewExceptionsApp(repos.Exceptions, repos.Audits, clk)
	summaryApp := application.NewSummaryApp(repos.Cycles, repos.Entries, repos.Exceptions, repos.Rules, repos.Summaries, repos.Recalcs, repos.Annuals, repos.Audits, clk)
	workspaceApp := application.NewWorkspaceApp(repos.Exceptions, nilNotifyAdapter{}, clk)
	auditApp := application.NewAuditApp(repos.Audits, repos.Exceptions, repos.Entries, clk)
	usersApp := application.NewUsersApp(repos.Users, repos.Audits, clk)
	deps := Router{
		Projects: projectsApp, Parties: partiesApp, Batches: batchesApp, Cycles: cyclesApp,
		Rules: rulesApp, Imports: importsApp, Exceptions: exceptionsApp, Summary: summaryApp,
		Workspace: workspaceApp, Audit: auditApp, Users: usersApp, Evaluate: evaluateApp,
	}
	return deps, repos, clk
}

type nilNotifyAdapter struct{}

func (nilNotifyAdapter) Send(ctx context.Context, recipient, channel, subject, body string) error {
	return nil
}

func TestHealthz(t *testing.T) {
	deps, _, _ := newTestRouter(t)
	engine := New(deps)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "ok") {
		t.Error("expected body to contain 'ok'")
	}
}

func TestReadyz(t *testing.T) {
	deps, _, _ := newTestRouter(t)
	engine := New(deps)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateProjectAuth(t *testing.T) {
	deps, _, _ := newTestRouter(t)
	engine := New(deps)
	body := `{"code":"WS-01","name":"P1","sponsor":"S","annual_budget_cents":1000,"start_year":2026,"end_year":2027}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("X-Tenant-ID", "t1")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without actor, got %d", w.Code)
	}
}

func TestCreateProjectAsAdmin(t *testing.T) {
	deps, _, _ := newTestRouter(t)
	engine := New(deps, WithDefaultTenant("t1"), WithActor("admin", Actor{
		UserID: "admin", Username: "admin", Role: domain.RoleAdmin, TenantID: "t1",
	}))
	body := `{"code":"WS-01","name":"P1","sponsor":"S","annual_budget_cents":1000,"start_year":2026,"end_year":2027}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("X-Tenant-ID", "t1")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["code"] != "WS-01" {
		t.Errorf("expected code WS-01, got %v", resp["code"])
	}
}

func TestCreateProjectAsOperator(t *testing.T) {
	deps, _, _ := newTestRouter(t)
	engine := New(deps, WithDefaultTenant("t1"), WithActor("op", Actor{
		UserID: "op", Username: "op", Role: domain.RoleOperator, TenantID: "t1",
	}))
	body := `{"code":"WS-01","name":"P1","sponsor":"S","annual_budget_cents":1000,"start_year":2026,"end_year":2027}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer op")
	req.Header.Set("X-Tenant-ID", "t1")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for operator, got %d", w.Code)
	}
}

func TestListProjects(t *testing.T) {
	deps, repos, clk := newTestRouter(t)
	ctx := context.Background()
	_, _ = repos.Projects.Create(ctx, domain.Project{
		ID: "p1", TenantID: "t1", Code: "WS-01", Name: "P1", Sponsor: "S",
		AnnualBudget: 1000, StartYear: 2026, EndYear: 2027,
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	})
	engine := New(deps, WithDefaultTenant("t1"), WithActor("op", Actor{
		UserID: "op", Username: "op", Role: domain.RoleOperator, TenantID: "t1",
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	req.Header.Set("Authorization", "Bearer op")
	req.Header.Set("X-Tenant-ID", "t1")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	items, ok := resp["items"].([]any)
	if !ok || len(items) != 1 {
		t.Errorf("expected 1 item, got %v", resp["items"])
	}
}

func TestCreateParty(t *testing.T) {
	deps, _, _ := newTestRouter(t)
	engine := New(deps, WithDefaultTenant("t1"), WithActor("op", Actor{
		UserID: "op", Username: "op", Role: domain.RoleOperator, TenantID: "t1",
	}))
	body := `{"name":"Donor","type":"donor","contact":"donor@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/parties", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer op")
	req.Header.Set("X-Tenant-ID", "t1")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestImportEntries(t *testing.T) {
	deps, repos, clk := newTestRouter(t)
	ctx := context.Background()
	_, _ = repos.Projects.Create(ctx, domain.Project{
		ID: "p1", TenantID: "t1", Code: "WS-01", Name: "P1", Sponsor: "S",
		AnnualBudget: 1000, StartYear: 2026, EndYear: 2027,
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	})
	_, _ = repos.Batches.Create(ctx, domain.FundingBatch{
		ID: "b1", TenantID: "t1", ProjectID: "p1", Code: "FB-01",
		DonorPartyID: "d1", ImplementerPartyID: "i1",
		TotalAmount: 1000_00, Currency: "CNY", DisbursedAt: clk.Now(),
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	})
	_, _ = repos.Cycles.Create(ctx, domain.SettlementCycle{
		ID: "c1", TenantID: "t1", ProjectID: "p1", FundingBatchID: "b1",
		Year: 2026, Quarter: 1, StartDate: clk.Now(), EndDate: clk.Now().AddDate(0, 3, 0),
		CreatedAt: clk.Now(), UpdatedAt: clk.Now(),
	})
	engine := New(deps, WithDefaultTenant("t1"), WithActor("op", Actor{
		UserID: "op", Username: "op", Role: domain.RoleOperator, TenantID: "t1",
	}))
	ts := clk.Now().Format(time.RFC3339)
	body := `{"batch_id":"b1","cycle_id":"c1","project_id":"p1","entries":[{"source_id":"s1","source":"import","payee_party_id":"py1","payer_party_id":"pp1","amount_cents":100,"currency":"CNY","occurred_at":"` + ts + `"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/entries/import", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer op")
	req.Header.Set("X-Tenant-ID", "t1")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestAuditExportCSV(t *testing.T) {
	deps, repos, clk := newTestRouter(t)
	ctx := context.Background()
	_, _ = repos.Audits.Append(ctx, domain.AuditEntry{
		ID: "a1", TenantID: "t1", ActorID: "op", Action: domain.AuditActionImport,
		EntityType: "settlement_entry", EntityID: "e1", CreatedAt: clk.Now(),
	})
	engine := New(deps, WithDefaultTenant("t1"), WithActor("admin", Actor{
		UserID: "admin", Username: "admin", Role: domain.RoleAdmin, TenantID: "t1",
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/export", nil)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("X-Tenant-ID", "t1")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "a1") {
		t.Error("expected CSV to contain audit id a1")
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "text/csv") {
		t.Error("expected text/csv content type")
	}
}

func TestValidationError(t *testing.T) {
	deps, _, _ := newTestRouter(t)
	engine := New(deps, WithDefaultTenant("t1"), WithActor("admin", Actor{
		UserID: "admin", Username: "admin", Role: domain.RoleAdmin, TenantID: "t1",
	}))
	// Missing required fields.
	body := `{"code":"","name":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("X-Tenant-ID", "t1")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestParseRFC3339(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		got, err := parseRFC3339("2026-01-15T10:00:00Z")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got.Year() != 2026 {
			t.Errorf("expected 2026, got %d", got.Year())
		}
	})
	t.Run("empty", func(t *testing.T) {
		_, err := parseRFC3339("")
		if err == nil {
			t.Fatal("expected error for empty string")
		}
	})
	t.Run("invalid", func(t *testing.T) {
		_, err := parseRFC3339("not-a-date")
		if err == nil {
			t.Fatal("expected error for invalid string")
		}
	})
}

func TestFormatRFC3339(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		if got := formatRFC3339(time.Time{}); got != "" {
			t.Errorf("expected empty for zero time, got %s", got)
		}
	})
	t.Run("valid", func(t *testing.T) {
		ts := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
		got := formatRFC3339(ts)
		if got == "" {
			t.Error("expected non-empty for valid time")
		}
	})
}

func TestBindAndValidateJSONError(t *testing.T) {
	deps, _, _ := newTestRouter(t)
	engine := New(deps, WithDefaultTenant("t1"), WithActor("admin", Actor{
		UserID: "admin", Username: "admin", Role: domain.RoleAdmin, TenantID: "t1",
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader([]byte("not json")))
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("X-Tenant-ID", "t1")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
