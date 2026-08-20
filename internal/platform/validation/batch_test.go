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
