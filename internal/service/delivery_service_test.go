package service

import (
	"context"
	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/platform/clock"
	"github.com/wyw14/cry-046/internal/platform/ids"
	"github.com/wyw14/cry-046/internal/platform/notify"
	"github.com/wyw14/cry-046/internal/repository/memory"
	"testing"
	"time"
)

func TestDeliveryRequiresApprovedPalette(t *testing.T) {
	s := memory.NewStore()
	ps := NewProjectService(s, s, clock.System{}, &ids.Sequence{})
	_, _ = ps.Create(context.Background(), application.CreateProjectInput{ID: "p", Name: "n", OwnerID: "o"})
	pSvc := NewPaletteService(s, s, s, s, clock.System{}, &ids.Sequence{})
	p, _ := pSvc.Create(context.Background(), application.CreatePaletteInput{ID: "pa", ProjectID: "p", Name: "v", Entries: []application.ColorInput{{Name: "a", Hex: "#112233"}}})
	d := NewDeliveryService(s, s, s, s, LocalPackageWriter{Root: "tmp"}, s, notify.New(), clock.System{}, &ids.Sequence{})
	_, err := d.Request(context.Background(), application.CreateDeliveryInput{ID: "d", PaletteID: p.ID, Format: "json", ExpiresAt: time.Now().Add(time.Hour).Format(time.RFC3339)})
	if err == nil {
		t.Fatal("draft palette delivery accepted")
	}
}
