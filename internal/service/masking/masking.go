// Package masking provides field-level masking helpers for sensitive
// data so logs and API responses never leak contact info.
package masking

import (
	"strings"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// MaskParty returns a copy of the party with the contact field masked.
func MaskParty(p domain.Party) domain.Party {
	cp := p
	cp.Contact = domain.MaskContact(p.Contact)
	return cp
}

// MaskUser returns a copy of the user with the email field masked
// and the password hash wiped.
func MaskUser(u domain.User) domain.User {
	cp := u
	cp.Email = domain.MaskEmail(u.Email)
	cp.PasswordHash = ""
	return cp
}

// MaskParties masks a slice in place.
func MaskParties(in []domain.Party) []domain.Party {
	out := make([]domain.Party, len(in))
	for i, p := range in {
		out[i] = MaskParty(p)
	}
	return out
}

// MaskUsers masks a slice in place.
func MaskUsers(in []domain.User) []domain.User {
	out := make([]domain.User, len(in))
	for i, u := range in {
		out[i] = MaskUser(u)
	}
	return out
}

// MaskString masks the middle of any string, leaving the first
// and last char visible when len > 2.
func MaskString(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 2 {
		return strings.Repeat("*", len(s))
	}
	return s[:1] + strings.Repeat("*", len(s)-2) + s[len(s)-1:]
}
