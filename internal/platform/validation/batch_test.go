package validation

import (
	"github.com/wyw14/cry-046/internal/domain"
	"testing"
)

func TestPreflightDetectsDuplicateAndExisting(t *testing.T) {
	r := Preflight([]domain.ColorEntry{{Name: "主色", Hex: "#112233"}, {Name: "重复", Hex: "#112233"}, {Name: "已有", Hex: "#445566"}}, map[string]bool{"#445566": true})
	if len(r.Issues) != 2 || len(r.Accepted) != 1 {
		t.Fatalf("unexpected %#v", r)
	}
}

// Colors that differ only by letter case must be treated as the same color
// for both within-batch duplicates and already-existing colors.
func TestPreflightCaseInsensitiveHex(t *testing.T) {
	rows := []domain.ColorEntry{
		{Name: "红", Hex: "#FF0000"},
		{Name: "红2", Hex: "#ff0000"}, // duplicate of row 1, case-only difference
		{Name: "蓝", Hex: "#00AAFF"},
		{Name: "蓝2", Hex: "#00aaff"}, // already known (case-only difference)
	}
	r := Preflight(rows, map[string]bool{"#00AAFF": true})

	if len(r.Accepted) != 1 || r.Accepted[0].Hex != "#FF0000" {
		t.Fatalf("expected only #FF0000 accepted, got %#v", r.Accepted)
	}
	codes := map[string]bool{}
	for _, is := range r.Issues {
		codes[is.Code] = true
	}
	if !codes["DUPLICATE"] || !codes["EXISTS"] {
		t.Fatalf("expected DUPLICATE and EXISTS issues, got %#v", r.Issues)
	}
	if len(r.DuplicateHex) != 1 || r.DuplicateHex[0] != "#ff0000" {
		t.Fatalf("expected canonicalized duplicate hex, got %#v", r.DuplicateHex)
	}
}
