package domain

import (
	"strings"
	"time"
)

// Role is the role of an operator in the platform.
type Role string

const (
	RoleOperator Role = "operator" // can import, assign, comment
	RoleAssignee Role = "assignee" // can claim, process, submit for review
	RoleReviewer Role = "reviewer" // can resolve, close, rework
	RoleAdmin    Role = "admin"    // can configure rules, archive
)

// IsValidRole reports whether r is a known role.
func IsValidRole(r Role) bool {
	switch r {
	case RoleOperator, RoleAssignee, RoleReviewer, RoleAdmin:
		return true
	}
	return false
}

// User is the operator record. Auth is local-only: a hashed password
// and a session token. The platform does not call any external IdP.
type User struct {
	ID           string
	TenantID     string
	Username     string
	DisplayName  string
	Email        string
	Role         Role
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Validate checks invariants.
func (u User) Validate() error {
	if u.ID == "" {
		return NewErr(CodeInvalidArgument, "user id must not be empty").WithField("id")
	}
	if strings.TrimSpace(u.Username) == "" {
		return NewErr(CodeInvalidArgument, "username must not be empty").WithField("username")
	}
	if !IsValidRole(u.Role) {
		return NewErr(CodeInvalidArgument, "invalid role").WithField("role")
	}
	if strings.TrimSpace(u.Email) != "" && !strings.Contains(u.Email, "@") {
		return NewErr(CodeInvalidArgument, "invalid email").WithField("email")
	}
	return nil
}

// Can reports whether the user has the given role.
func (u User) Can(r Role) bool { return u.Role == r }

// CanResolve reports whether the user can resolve (must be reviewer or admin).
func (u User) CanResolve() bool { return u.Role == RoleReviewer || u.Role == RoleAdmin }

// CanAssign reports whether the user can assign.
func (u User) CanAssign() bool {
	switch u.Role {
	case RoleOperator, RoleReviewer, RoleAdmin:
		return true
	}
	return false
}

// CanArchive reports whether the user can archive rule versions.
func (u User) CanArchive() bool { return u.Role == RoleAdmin }

// MaskEmail masks the local part of an email for safe display.
func MaskEmail(s string) string {
	if !strings.Contains(s, "@") {
		return s
	}
	parts := strings.SplitN(s, "@", 2)
	local := parts[0]
	if len(local) <= 2 {
		return strings.Repeat("*", len(local)) + "@" + parts[1]
	}
	return local[:1] + strings.Repeat("*", len(local)-2) + local[len(local)-1:] + "@" + parts[1]
}

// AuditAction is the high-level action recorded in the audit log.
type AuditAction string

const (
	AuditActionImport      AuditAction = "import"
	AuditActionAssign      AuditAction = "assign"
	AuditActionClaim       AuditAction = "claim"
	AuditActionResubmit    AuditAction = "resubmit"
	AuditActionReview      AuditAction = "review"
	AuditActionResolve     AuditAction = "resolve"
	AuditActionClose       AuditAction = "close"
	AuditActionEscalate    AuditAction = "escalate"
	AuditActionRework      AuditAction = "rework"
	AuditActionComment     AuditAction = "comment"
	AuditActionAttach      AuditAction = "attach"
	AuditActionRulePublish AuditAction = "rule_publish"
	AuditActionRuleArchive AuditAction = "rule_archive"
	AuditActionRecalculate AuditAction = "recalculate"
	AuditActionExport      AuditAction = "export"
	AuditActionArchive     AuditAction = "archive"
	AuditActionLogin       AuditAction = "login"
	AuditActionLogout      AuditAction = "logout"
)

// AuditEntry is the immutable audit record.
type AuditEntry struct {
	ID         string
	TenantID   string
	ActorID    string
	Action     AuditAction
	EntityType string
	EntityID   string
	Detail     map[string]string
	CreatedAt  time.Time
}

// Validate checks invariants.
func (a AuditEntry) Validate() error {
	if a.ID == "" {
		return NewErr(CodeInvalidArgument, "audit id must not be empty").WithField("id")
	}
	if a.ActorID == "" {
		return NewErr(CodeInvalidArgument, "actor id must not be empty").WithField("actor_id")
	}
	if strings.TrimSpace(string(a.Action)) == "" {
		return NewErr(CodeInvalidArgument, "action must not be empty").WithField("action")
	}
	if a.CreatedAt.IsZero() {
		return NewErr(CodeInvalidArgument, "created_at must not be empty").WithField("created_at")
	}
	return nil
}
