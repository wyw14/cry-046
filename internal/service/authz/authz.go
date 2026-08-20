// Package authz provides a small RBAC layer used by the transport
// middleware to enforce role boundaries on API endpoints. The rules
// are encoded as a map from action to required roles.
package authz

import (
	"github.com/welfare/settlement-resolver/internal/domain"
)

// Permission is a stable permission identifier.
type Permission string

const (
	PermProjectCreate     Permission = "project.create"
	PermProjectRead       Permission = "project.read"
	PermPartyCreate       Permission = "party.create"
	PermBatchCreate       Permission = "batch.create"
	PermCycleCreate       Permission = "cycle.create"
	PermRuleCreate        Permission = "rule.create"
	PermRulePublish       Permission = "rule.publish"
	PermRuleArchive       Permission = "rule.archive"
	PermEntryImport       Permission = "entry.import"
	PermEntryRead         Permission = "entry.read"
	PermExceptionList     Permission = "exception.list"
	PermExceptionAssign   Permission = "exception.assign"
	PermExceptionClaim    Permission = "exception.claim"
	PermExceptionReview   Permission = "exception.review"
	PermExceptionResolve  Permission = "exception.resolve"
	PermExceptionClose    Permission = "exception.close"
	PermExceptionEscalate Permission = "exception.escalate"
	PermExceptionRework   Permission = "exception.rework"
	PermExceptionComment  Permission = "exception.comment"
	PermExceptionAttach   Permission = "exception.attach"
	PermSummaryRead       Permission = "summary.read"
	PermSummaryRecalc     Permission = "summary.recalc"
	PermAuditRead         Permission = "audit.read"
	PermAuditExport       Permission = "audit.export"
	PermUserCreate        Permission = "user.create"
	PermWorkspaceRead     Permission = "workspace.read"
)

// AllowedRoles returns the set of roles that may perform the given
// permission. An empty slice means "no one"; the middleware maps this
// to a permission-denied error.
func AllowedRoles(p Permission) []domain.Role {
	switch p {
	case PermProjectCreate, PermBatchCreate, PermCycleCreate, PermRuleCreate, PermUserCreate:
		return []domain.Role{domain.RoleAdmin}
	case PermRulePublish, PermRuleArchive:
		return []domain.Role{domain.RoleAdmin}
	case PermPartyCreate:
		return []domain.Role{domain.RoleAdmin, domain.RoleOperator}
	case PermEntryImport:
		return []domain.Role{domain.RoleAdmin, domain.RoleOperator}
	case PermExceptionAssign:
		return []domain.Role{domain.RoleAdmin, domain.RoleOperator, domain.RoleReviewer}
	case PermExceptionResolve:
		return []domain.Role{domain.RoleAdmin, domain.RoleReviewer}
	case PermExceptionClose, PermExceptionRework:
		return []domain.Role{domain.RoleAdmin, domain.RoleReviewer}
	case PermExceptionEscalate:
		return []domain.Role{domain.RoleAdmin, domain.RoleAssignee, domain.RoleOperator}
	case PermExceptionReview, PermExceptionClaim, PermExceptionComment, PermExceptionAttach:
		return []domain.Role{domain.RoleAdmin, domain.RoleAssignee, domain.RoleOperator, domain.RoleReviewer}
	case PermSummaryRecalc:
		return []domain.Role{domain.RoleAdmin, domain.RoleOperator}
	case PermAuditExport:
		return []domain.Role{domain.RoleAdmin}
	}
	// Default: read is allowed for any authenticated role.
	return []domain.Role{domain.RoleAdmin, domain.RoleOperator, domain.RoleAssignee, domain.RoleReviewer}
}

// Can reports whether the user has permission p.
func Can(u domain.User, p Permission) bool {
	allowed := AllowedRoles(p)
	for _, r := range allowed {
		if u.Role == r {
			return true
		}
	}
	return false
}
