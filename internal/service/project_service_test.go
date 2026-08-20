package service

import (
	"context"
	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/domain"
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
	// Owner must still be able to archive, and state/audit must not change for non-owner.
	got, err := svc.Get(context.Background(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != domain.ProjectActive {
		t.Fatalf("non-owner archive changed state to %v", got.State)
	}
	if _, err = svc.Archive(context.Background(), p.ID, "owner"); err != nil {
		t.Fatalf("owner archive failed: %v", err)
	}
	got, _ = svc.Get(context.Background(), p.ID)
	if got.State != domain.ProjectArchived {
		t.Fatalf("owner archive did not change state: %v", got.State)
	}
}
