package memory

import (
	"context"
	"github.com/wyw14/cry-046/internal/domain"
	"testing"
)

func TestStoreOptimisticProjectUpdate(t *testing.T) {
	s := NewStore()
	p, _ := domain.NewProject("p", "n", "u")
	if err := s.CreateProject(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	p.Name = "new"
	if err := s.UpdateProject(context.Background(), p, 99); err != ErrConflict {
		t.Fatalf("want conflict got %v", err)
	}
}
