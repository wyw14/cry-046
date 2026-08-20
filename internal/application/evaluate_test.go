package application

import (
	"context"
	"strings"
	"testing"

	"github.com/welfare/settlement-resolver/internal/domain"
)

func TestEvaluateApp_Idempotency(t *testing.T) {
	clk := newFakeClock()
	now := clk.Now()

	// Build a published rule version with one matching rule.
	rv := domain.RuleVersion{
		ID: "rv-1", TenantID: "t1", Code: "RV-1", ProjectID: "p1",
		Description: "test",
		Rules: []domain.RuleDefinition{
			{ID: "r1", Code: "AMOUNT_ZERO", Severity: domain.SeverityHigh, Expression: "amount == 0", DeadlineHours: 48},
		},
		Status:      domain.RuleVersionStatusPublished,
		PublishedAt: now,
		Version:     1,
	}
	rules := &ruleVersionRepoFake{rv: rv}

	// Entries: one with amount 0 (matches), one with amount 100 (does not match).
	entries := &entryRepoFake{}
	entries.entries = []domain.SettlementEntry{
		{
			ID: "e1", TenantID: "t1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
			Source: domain.EntrySourceImport, PayeePartyID: "py", PayerPartyID: "pp",
			Amount: 0, Currency: "CNY", OccurredAt: now,
			SourceFingerprint: "fp1",
		},
		{
			ID: "e2", TenantID: "t1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
			Source: domain.EntrySourceImport, PayeePartyID: "py", PayerPartyID: "pp",
			Amount: 100, Currency: "CNY", OccurredAt: now,
			SourceFingerprint: "fp2",
		},
	}

	excs := &exceptionRepoFake{}
	audit := &auditRepoFake{}

	app := NewEvaluateApp(rules, entries, excs, audit, clk)

	res1, err := app.EvaluateCycle(context.Background(), EvaluateCycleInput{
		TenantID: "t1", CycleID: "c1", RuleVersionID: "rv-1", ActorID: "admin",
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if res1.CreatedExceptions != 1 {
		t.Fatalf("expected 1 created, got %d", res1.CreatedExceptions)
	}
	if res1.HitEntries != 1 {
		t.Fatalf("expected 1 hit entry, got %d", res1.HitEntries)
	}
	if res1.ScannedEntries != 2 {
		t.Fatalf("expected 2 scanned, got %d", res1.ScannedEntries)
	}

	// Re-evaluating with the same rule version must NOT create new exceptions.
	res2, err := app.EvaluateCycle(context.Background(), EvaluateCycleInput{
		TenantID: "t1", CycleID: "c1", RuleVersionID: "rv-1", ActorID: "admin",
	})
	if err != nil {
		t.Fatalf("second evaluate: %v", err)
	}
	if res2.CreatedExceptions != 0 {
		t.Fatalf("expected 0 created on re-evaluation, got %d", res2.CreatedExceptions)
	}
}

func TestEvaluateApp_UnpublishedRulesRejected(t *testing.T) {
	clk := newFakeClock()
	rv := domain.RuleVersion{
		ID: "rv-1", TenantID: "t1", Code: "RV-1", ProjectID: "p1",
		Rules:   []domain.RuleDefinition{{ID: "r1", Code: "X", Severity: domain.SeverityLow, Expression: "amount == 0"}},
		Status:  domain.RuleVersionStatusDraft,
		Version: 1,
	}
	rules := &ruleVersionRepoFake{rv: rv}
	entries := &entryRepoFake{}
	excs := &exceptionRepoFake{}
	audit := &auditRepoFake{}
	app := NewEvaluateApp(rules, entries, excs, audit, clk)

	_, err := app.EvaluateCycle(context.Background(), EvaluateCycleInput{
		TenantID: "t1", CycleID: "c1", RuleVersionID: "rv-1", ActorID: "admin",
	})
	if err == nil || !strings.Contains(err.Error(), "not published") {
		t.Fatalf("expected 'not published' error, got %v", err)
	}
}

func TestEvaluateApp_SnapshotCaptured(t *testing.T) {
	clk := newFakeClock()
	now := clk.Now()
	rv := domain.RuleVersion{
		ID: "rv-1", TenantID: "t1", Code: "RV-1", ProjectID: "p1",
		Rules: []domain.RuleDefinition{
			{ID: "r1", Code: "AMOUNT_ZERO", Severity: domain.SeverityHigh, Expression: "amount == 0"},
		},
		Status:      domain.RuleVersionStatusPublished,
		PublishedAt: now,
		Version:     1,
	}
	rules := &ruleVersionRepoFake{rv: rv}
	entries := &entryRepoFake{
		entries: []domain.SettlementEntry{
			{
				ID: "e1", TenantID: "t1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
				Source: domain.EntrySourceImport, PayeePartyID: "py", PayerPartyID: "pp",
				Amount: 0, Currency: "CNY", OccurredAt: now,
				SourceFingerprint: "fp1",
				Metadata:          map[string]string{"k": "v"},
			},
		},
	}
	excs := &exceptionRepoFake{}
	audit := &auditRepoFake{}
	app := NewEvaluateApp(rules, entries, excs, audit, clk)

	_, err := app.EvaluateCycle(context.Background(), EvaluateCycleInput{
		TenantID: "t1", CycleID: "c1", RuleVersionID: "rv-1", ActorID: "admin",
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(excs.items) != 1 {
		t.Fatalf("expected 1 exception, got %d", len(excs.items))
	}
	snap := excs.items[0].Snapshot
	if snap.EntryAmountCents != 0 {
		t.Errorf("snapshot amount: %d", snap.EntryAmountCents)
	}
	if snap.EntryCurrency != "CNY" {
		t.Errorf("snapshot currency: %s", snap.EntryCurrency)
	}
	if snap.RuleExpression != "amount == 0" {
		t.Errorf("snapshot rule expression: %s", snap.RuleExpression)
	}
	if snap.SnapshotAt.IsZero() {
		t.Error("snapshot time must be set")
	}
	if snap.InputFields["k"] != "v" {
		t.Errorf("snapshot input fields: %+v", snap.InputFields)
	}
}
