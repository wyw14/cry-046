package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
)

func TestAuthorizationRequiresAdminAndTenantMatch(t *testing.T) {
	reviewer := domain.User{ID: "u1", TenantID: "t1", Role: domain.RoleReviewer}
	got, err := EvaluateAuthorization(context.Background(), reviewer, "t2", []string{"archive_rule", "resolve_exception"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Allowed != 0 {
		t.Fatalf("reviewer/cross-tenant actions allowed: %+v", got)
	}
}
