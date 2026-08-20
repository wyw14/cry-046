package domain

import (
	"testing"
	"time"
)

func TestProjectArchiveRequiresActive(t *testing.T) {
	p, _ := NewProject("p1", "方案", "u1")
	if err := p.Archive(time.Unix(1, 0)); err != nil {
		t.Fatal(err)
	}
	if p.CanEdit() {
		t.Fatal("archived project editable")
	}
	if err := p.Archive(time.Unix(2, 0)); err == nil {
		t.Fatal("double archive allowed")
	}
}
