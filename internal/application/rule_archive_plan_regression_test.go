package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
	"time"
)

func TestRuleArchivePlanRetainsPublicationAuditIdentity(t *testing.T) {
	published := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	archivedAt := published.Add(24 * time.Hour)
	versions := []domain.RuleVersion{{ID: "r2", TenantID: "t1", ProjectID: "p1", Code: "B", Status: domain.RuleVersionStatusPublished, PublishedAt: published, Rules: []domain.RuleDefinition{{ID: "x", Code: "X"}}}, {ID: "r1", TenantID: "t1", ProjectID: "p1", Code: "A", Status: domain.RuleVersionStatusPublished, PublishedAt: published.Add(-time.Hour), Rules: []domain.RuleDefinition{{ID: "y", Code: "Y"}}}}
	got, err := BuildRuleArchivePlan(context.Background(), "t1", versions, archivedAt)
	if err != nil {
		t.Fatal(err)
	}
	if got.Archived[0].PublishedAt != published || got.Records[0].PublishedAt.IsZero() {
		t.Fatalf("publication audit identity lost: %+v", got)
	}
	if got.Records[0].Code != "A" || got.Records[0].Sequence != 1 {
		t.Fatalf("records not deterministic: %+v", got.Records)
	}
}
