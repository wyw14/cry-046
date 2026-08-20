package service

import (
	"context"
	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/platform/clock"
	"github.com/wyw14/cry-046/internal/platform/ids"
	"github.com/wyw14/cry-046/internal/repository/memory"
	"testing"
)

func TestAssetExportRequiresCopyright(t *testing.T) {
	s := memory.NewStore()
	ps := NewProjectService(s, s, clock.System{}, &ids.Sequence{})
	if _, err := ps.Create(context.Background(), application.CreateProjectInput{ID: "p", Name: "方案", OwnerID: "o"}); err != nil {
		t.Fatal(err)
	}
	as := NewAssetService(s, s, s, clock.System{}, &ids.Sequence{})
	a, err := as.Create(context.Background(), application.CreateAssetInput{ID: "a", ProjectID: "p", Name: "素材", Filename: "a.png", Mime: "image/png", Bytes: 12})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := as.Exportable(context.Background(), a.ID, clock.System{}.Now()); err == nil {
		t.Fatal("asset without copyright exported")
	}
}
