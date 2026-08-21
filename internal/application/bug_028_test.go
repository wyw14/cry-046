package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
)

func TestProjectsList_NormalizesPageNumber(t *testing.T) {
	r := &recordProjectRepo{}
	app := NewProjectsApp(r, nil, nil, nil, nil, newFakeClock())
	_, _, err := app.List(context.Background(), ListQuery{TenantID: "t1", Page: 0, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if r.q.Page != 1 {
		t.Fatalf("page was not normalized: %d", r.q.Page)
	}
}

type recordProjectRepo struct{ q ListQuery }

func (r *recordProjectRepo) Create(context.Context, domain.Project) (domain.Project, error) {
	return domain.Project{}, nil
}
func (r *recordProjectRepo) Get(context.Context, string, string) (domain.Project, error) {
	return domain.Project{}, nil
}
func (r *recordProjectRepo) Update(context.Context, domain.Project) (domain.Project, error) {
	return domain.Project{}, nil
}
func (r *recordProjectRepo) List(_ context.Context, q ListQuery) ([]domain.Project, PageResult, error) {
	r.q = q
	return nil, PageResult{}, nil
}
