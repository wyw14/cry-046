package service

import (
	"context"
	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/platform/clock"
	"github.com/wyw14/cry-046/internal/platform/ids"
	"github.com/wyw14/cry-046/internal/repository/memory"
	"testing"
)

func TestWorkspaceRecentDeduplicatesProject(t *testing.T) {
	s := memory.NewStore()
	p := NewProjectService(s, s, clock.System{}, &ids.Sequence{})
	_, _ = p.Create(context.Background(), application.CreateProjectInput{ID: "p", Name: "n", OwnerID: "o"})
	w := NewWorkspace(s, clock.System{}, &ids.Sequence{})
	_ = w.RecordVisit(context.Background(), "u", "p")
	_ = w.RecordVisit(context.Background(), "u", "p")
	items, _ := w.ListRecent(context.Background(), "u", 10)
	if len(items) != 1 {
		t.Fatalf("got %d visits", len(items))
	}
}
