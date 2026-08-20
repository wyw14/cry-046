package authz

import (
	"testing"

	"github.com/welfare/settlement-resolver/internal/domain"
)

func TestAllowedRoles(t *testing.T) {
	cases := []struct {
		perm  Permission
		roles []domain.Role
		want  bool
	}{
		{PermProjectCreate, []domain.Role{domain.RoleAdmin}, true},
		{PermProjectCreate, []domain.Role{domain.RoleOperator}, false},
		{PermPartyCreate, []domain.Role{domain.RoleOperator}, true},
		{PermPartyCreate, []domain.Role{domain.RoleAssignee}, false},
		{PermExceptionResolve, []domain.Role{domain.RoleReviewer}, true},
		{PermExceptionResolve, []domain.Role{domain.RoleAssignee}, false},
		{PermExceptionClaim, []domain.Role{domain.RoleAssignee}, true},
		{PermEntryImport, []domain.Role{domain.RoleOperator}, true},
	}
	for _, tc := range cases {
		t.Run(string(tc.perm)+"/"+string(tc.roles[0]), func(t *testing.T) {
			allowed := AllowedRoles(tc.perm)
			found := false
			for _, r := range allowed {
				if r == tc.roles[0] {
					found = true
					break
				}
			}
			if found != tc.want {
				t.Errorf("expected %v, got %v", tc.want, found)
			}
		})
	}
}

func TestCan(t *testing.T) {
	admin := domain.User{Role: domain.RoleAdmin}
	if !Can(admin, PermProjectCreate) {
		t.Error("admin should be able to create projects")
	}
	if !Can(admin, PermRuleArchive) {
		t.Error("admin should be able to archive rules")
	}
	assignee := domain.User{Role: domain.RoleAssignee}
	if Can(assignee, PermProjectCreate) {
		t.Error("assignee should not be able to create projects")
	}
	if !Can(assignee, PermExceptionClaim) {
		t.Error("assignee should be able to claim exceptions")
	}
}
