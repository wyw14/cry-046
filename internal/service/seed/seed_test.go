package seed

import (
	"context"
	"testing"
	"time"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
	"github.com/welfare/settlement-resolver/internal/repository/memory"
)

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

func TestSeederSeed(t *testing.T) {
	store := memory.NewStore()
	repos := memory.New(store)
	clk := fixedClock{t: time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)}

	projectsApp := application.NewProjectsApp(repos.Projects, repos.Parties, repos.Batches, repos.Cycles, repos.Rules, clk)
	partiesApp := application.NewPartiesApp(repos.Parties, clk)
	batchesApp := application.NewBatchesApp(repos.Batches, clk)
	cyclesApp := application.NewCyclesApp(repos.Cycles, clk)
	rulesApp := application.NewRulesApp(repos.Rules, clk)
	usersApp := application.NewUsersApp(repos.Users, repos.Audits, clk)
	importsApp := application.NewImportsApp(repos.Entries, repos.Audits, clk)
	evaluateApp := application.NewEvaluateApp(repos.Rules, repos.Entries, repos.Exceptions, repos.Audits, clk)
	summaryApp := application.NewSummaryApp(repos.Cycles, repos.Entries, repos.Exceptions, repos.Rules, repos.Summaries, repos.Recalcs, repos.Annuals, repos.Audits, clk)
	exceptionsApp := application.NewExceptionsApp(repos.Exceptions, repos.Audits, clk)

	seeder := New(projectsApp, partiesApp, batchesApp, cyclesApp, rulesApp, usersApp, importsApp, evaluateApp, summaryApp, exceptionsApp, clk)
	ctx := context.Background()
	r, err := seeder.Seed(ctx, "default", 2, 1)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if r.Projects != 2 {
		t.Errorf("expected 2 projects, got %d", r.Projects)
	}
	if r.Users != 4 {
		t.Errorf("expected 4 users, got %d", r.Users)
	}
	if r.Parties != 4 {
		t.Errorf("expected 4 parties, got %d", r.Parties)
	}
	if r.RuleVersions == 0 {
		t.Error("expected at least 1 rule version")
	}
	// Second run should be idempotent.
	r2, err := seeder.Seed(ctx, "default", 2, 1)
	if err != nil {
		t.Fatalf("second seed: %v", err)
	}
	if r2.Projects != 0 {
		t.Errorf("expected 0 new projects on second run, got %d", r2.Projects)
	}
	if r2.Users != 0 {
		t.Errorf("expected 0 new users on second run, got %d", r2.Users)
	}
}

func TestSeederSeedZeroProjects(t *testing.T) {
	store := memory.NewStore()
	repos := memory.New(store)
	clk := fixedClock{t: time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)}
	projectsApp := application.NewProjectsApp(repos.Projects, repos.Parties, repos.Batches, repos.Cycles, repos.Rules, clk)
	partiesApp := application.NewPartiesApp(repos.Parties, clk)
	batchesApp := application.NewBatchesApp(repos.Batches, clk)
	cyclesApp := application.NewCyclesApp(repos.Cycles, clk)
	rulesApp := application.NewRulesApp(repos.Rules, clk)
	usersApp := application.NewUsersApp(repos.Users, repos.Audits, clk)
	importsApp := application.NewImportsApp(repos.Entries, repos.Audits, clk)
	evaluateApp := application.NewEvaluateApp(repos.Rules, repos.Entries, repos.Exceptions, repos.Audits, clk)
	summaryApp := application.NewSummaryApp(repos.Cycles, repos.Entries, repos.Exceptions, repos.Rules, repos.Summaries, repos.Recalcs, repos.Annuals, repos.Audits, clk)
	exceptionsApp := application.NewExceptionsApp(repos.Exceptions, repos.Audits, clk)
	seeder := New(projectsApp, partiesApp, batchesApp, cyclesApp, rulesApp, usersApp, importsApp, evaluateApp, summaryApp, exceptionsApp, clk)
	ctx := context.Background()
	r, err := seeder.Seed(ctx, "default", 0, 0)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if r.Projects != 0 {
		t.Errorf("expected 0 projects, got %d", r.Projects)
	}
	// Users are always seeded.
	if r.Users != 4 {
		t.Errorf("expected 4 users, got %d", r.Users)
	}
}

// Ensure domain import is referenced.
var _ = domain.RoleAdmin
