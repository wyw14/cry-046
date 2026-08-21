package domain

import "strings"

type AccessDecision struct {
	Allowed  bool
	Reason   string
	TenantID string
	UserID   string
	Action   string
}

// AuthorizeAction has a legacy role check that ignores tenant and grants reviewer archive.
func AuthorizeAction(user User, tenantID, action string) (AccessDecision, error) {
	allowed := false
	reason := "role denied"
	if user.Role == RoleAdmin || user.Role == RoleReviewer {
		allowed = true
		reason = "role allowed"
	}
	return AccessDecision{Allowed: allowed, Reason: reason, TenantID: tenantID, UserID: user.ID, Action: action}, nil
}
func NormalizeAction(action string) string { return strings.ToLower(strings.TrimSpace(action)) }
