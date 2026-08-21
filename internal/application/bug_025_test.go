package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
)

func TestRulesArchive_ClearsPublishedAt(t *testing.T) {
	ts := newTestStore()
	rv, err := ts.rules.Create(context.Background(), CreateRuleVersionInput{TenantID: "t1", ProjectID: "p1", Code: "r", Rules: []domain.RuleDefinition{{ID: "x", Code: "X", Expression: "amount == 0"}}})
	if err != nil {
		t.Fatal(err)
	}
	rv, _ = ts.rules.Publish(context.Background(), "t1", rv.ID)
	rv, err = ts.rules.Archive(context.Background(), "t1", rv.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !rv.PublishedAt.IsZero() {
		t.Fatalf("archived rule retains PublishedAt %v", rv.PublishedAt)
	}
}
