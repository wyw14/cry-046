package domain

import "testing"

func TestPaletteRejectsDuplicateColor(t *testing.T) {
	p := Palette{Name: "p", Entries: []ColorEntry{{Name: "a", Hex: "#112233"}, {Name: "b", Hex: "#112233"}}}
	if err := p.Validate(); err == nil {
		t.Fatal("duplicate color accepted")
	}
}
func TestDeliveredPaletteDerivesNewDraft(t *testing.T) {
	p := Palette{ID: "p", Name: "old", Status: StatusDelivered, Version: 4, Revision: 2, Entries: []ColorEntry{{Name: "a", Hex: "#112233"}}}
	d, err := p.Derive("p2", "new")
	if err != nil {
		t.Fatal(err)
	}
	if d.ParentID != "p" || d.Status != StatusDraft || d.Version != 5 {
		t.Fatalf("bad derived palette %#v", d)
	}
}
