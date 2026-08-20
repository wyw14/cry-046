package service

import (
	"context"
	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/platform/clock"
	"github.com/wyw14/cry-046/internal/platform/ids"
	"github.com/wyw14/cry-046/internal/repository/memory"
	"testing"
)

func TestPaletteServiceDeliveredIsImmutable(t *testing.T) {
	s := memory.NewStore()
	pSvc := NewProjectService(s, s, clock.System{}, &ids.Sequence{})
	_, _ = pSvc.Create(context.Background(), application.CreateProjectInput{ID: "p", Name: "方案", OwnerID: "o"})
	svc := NewPaletteService(s, s, s, s, clock.System{}, &ids.Sequence{})
	p, _ := svc.Create(context.Background(), application.CreatePaletteInput{ID: "p1", ProjectID: "p", Name: "v1", Entries: []application.ColorInput{{Name: "a", Hex: "#112233"}}})
	_, _ = svc.Submit(context.Background(), p.ID, "o")
	_, _ = svc.Approve(context.Background(), p.ID, "o")
	p, _ = svc.Deliver(context.Background(), p.ID, "o")
	if _, err := svc.Submit(context.Background(), p.ID, "o"); err == nil {
		t.Fatal("delivered palette accepted edit")
	}
}
