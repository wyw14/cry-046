package service

import (
	"context"
	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/platform/clock"
	"github.com/wyw14/cry-046/internal/platform/ids"
	"github.com/wyw14/cry-046/internal/repository/memory"
	"testing"
)

func TestProjectServiceArchiveOwnerOnly(t *testing.T) {
	s := memory.NewStore()
	svc := NewProjectService(s, s, clock.System{}, &ids.Sequence{})
	p, err := svc.Create(context.Background(), application.CreateProjectInput{ID: "p", Name: "方案", OwnerID: "owner"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Archive(context.Background(), p.ID, "other"); err != ErrForbidden {
		t.Fatalf("want forbidden got %v", err)
	}
}
