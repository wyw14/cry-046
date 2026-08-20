package domain

import (
	"regexp"
	"strings"
	"time"
)

// PartyType enumerates the kinds of counterparty that participate in
// a settlement. The platform only stores parties that are referenced
// by settlement entries; it does not act as a CRM.
type PartyType string

const (
	PartyTypeDonor        PartyType = "donor"
	PartyTypeImplementer  PartyType = "implementer"
	PartyTypeBeneficiary  PartyType = "beneficiary"
	PartyTypeIntermediary PartyType = "intermediary"
)

// Party is a counterparty profile. Names are non-empty and unique within
// a tenant; the TenantID is included so the platform can support multiple
// operators in the future even though the demo only uses one.
type Party struct {
	ID        string
	TenantID  string
	Name      string
	Type      PartyType
	Contact   string
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks the invariants for a Party.
func (p Party) Validate() error {
	if p.ID == "" {
		return NewErr(CodeInvalidArgument, "party id must not be empty").WithField("id")
	}
	if p.TenantID == "" {
		return NewErr(CodeInvalidArgument, "tenant id must not be empty").WithField("tenant_id")
	}
	if strings.TrimSpace(p.Name) == "" {
		return NewErr(CodeInvalidArgument, "party name must not be empty").WithField("name")
	}
	if !isValidPartyType(p.Type) {
		return NewErr(CodeInvalidArgument, "invalid party type").WithField("type")
	}
	return nil
}

func isValidPartyType(t PartyType) bool {
	switch t {
	case PartyTypeDonor, PartyTypeImplementer, PartyTypeBeneficiary, PartyTypeIntermediary:
		return true
	}
	return false
}

// PartyContact is a value object representing the contact channel of a party.
var phoneRE = regexp.MustCompile(`^\+?[0-9]{6,15}$`)

// SanitiseContact strips whitespace and validates that the contact looks
// like an email address or phone number. This is a deliberate, conservative
// check rather than a permissive RFC parser. Local-only emails of the form
// "user@host" (without a dot) are accepted so the offline demo can use
// synthetic addresses like "admin@local".
func SanitiseContact(raw string) (string, error) {
	c := strings.TrimSpace(raw)
	if c == "" {
		return "", NewErr(CodeInvalidArgument, "contact must not be empty").WithField("contact")
	}
	if strings.Contains(c, "@") {
		parts := strings.SplitN(c, "@", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" || len(c) < 5 {
			return "", NewErr(CodeInvalidArgument, "invalid email").WithField("contact")
		}
		return c, nil
	}
	if phoneRE.MatchString(c) {
		return c, nil
	}
	return "", NewErr(CodeInvalidArgument, "invalid contact").WithField("contact")
}

// MaskContact masks the middle of a contact string for safe display.
func MaskContact(c string) string {
	if c == "" {
		return ""
	}
	if strings.Contains(c, "@") {
		parts := strings.SplitN(c, "@", 2)
		local := parts[0]
		if len(local) <= 2 {
			return strings.Repeat("*", len(local)) + "@" + parts[1]
		}
		return local[:1] + strings.Repeat("*", len(local)-2) + local[len(local)-1:] + "@" + parts[1]
	}
	if len(c) <= 4 {
		return strings.Repeat("*", len(c))
	}
	return c[:2] + strings.Repeat("*", len(c)-4) + c[len(c)-2:]
}
