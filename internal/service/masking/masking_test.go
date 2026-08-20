package masking

import (
	"testing"

	"github.com/welfare/settlement-resolver/internal/domain"
)

func TestMaskParty(t *testing.T) {
	p := domain.Party{ID: "p1", Contact: "donor@example.com"}
	masked := MaskParty(p)
	if masked.ID != "p1" {
		t.Error("ID should be preserved")
	}
	if masked.Contact == "donor@example.com" {
		t.Error("contact should be masked")
	}
	if masked.Contact == "" {
		t.Error("contact should not be empty")
	}
}

func TestMaskUser(t *testing.T) {
	u := domain.User{ID: "u1", Email: "admin@example.com", PasswordHash: "secret"}
	masked := MaskUser(u)
	if masked.ID != "u1" {
		t.Error("ID should be preserved")
	}
	if masked.Email == "admin@example.com" {
		t.Error("email should be masked")
	}
	if masked.PasswordHash != "" {
		t.Error("password hash should be wiped")
	}
}

func TestMaskParties(t *testing.T) {
	in := []domain.Party{
		{ID: "p1", Contact: "donor@example.com"},
		{ID: "p2", Contact: "+8613800000001"},
	}
	out := MaskParties(in)
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}
	for _, p := range out {
		if p.Contact == "donor@example.com" || p.Contact == "+8613800000001" {
			t.Error("contact should be masked")
		}
	}
}

func TestMaskUsers(t *testing.T) {
	in := []domain.User{
		{ID: "u1", Email: "admin@example.com", PasswordHash: "h1"},
		{ID: "u2", Email: "op@example.com", PasswordHash: "h2"},
	}
	out := MaskUsers(in)
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}
	for _, u := range out {
		if u.PasswordHash != "" {
			t.Error("password hash should be empty")
		}
	}
}

func TestMaskString(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"a", "*"},
		{"ab", "**"},
		{"abc", "a*c"},
		{"abcdef", "a****f"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := MaskString(tc.in); got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
