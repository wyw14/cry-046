package memory

import (
	"context"
	"sort"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// ProjectRepo is the in-memory ProjectRepo.
type ProjectRepo struct {
	store *Store
}

func (r *ProjectRepo) Create(ctx context.Context, p domain.Project) (domain.Project, error) {
	if err := p.Validate(); err != nil {
		return domain.Project{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	codeKey := p.TenantID + "|" + p.Code
	if _, ok := r.store.projectsByCode[codeKey]; ok {
		return domain.Project{}, domain.NewErrf(domain.CodeAlreadyExists, "project code %s already exists", p.Code).WithField("code")
	}
	r.store.projects[p.ID] = p
	r.store.projectsByCode[codeKey] = p.ID
	return p, nil
}

func (r *ProjectRepo) Get(ctx context.Context, tenantID, id string) (domain.Project, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	p, ok := r.store.projects[id]
	if !ok || p.TenantID != tenantID {
		return domain.Project{}, domain.NewErrf(domain.CodeNotFound, "project %s not found", id)
	}
	return p, nil
}

func (r *ProjectRepo) List(ctx context.Context, q application.ListQuery) ([]domain.Project, application.PageResult, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.Project, 0, len(r.store.projects))
	for _, p := range r.store.projects {
		if p.TenantID != q.TenantID {
			continue
		}
		if !matchesFilters(q.Filters, projectToFilterMap(p)) {
			continue
		}
		out = append(out, p)
	}
	sortProjects(out, q.OrderBy, q.OrderDesc)
	page := pageSlice(out, q.Page, q.PageSize)
	res := application.PageResult{
		Page: q.Page, PageSize: q.PageSize, Total: len(out),
		HasNext: q.Page*q.PageSize < len(out),
	}
	return page, res, nil
}

func (r *ProjectRepo) Update(ctx context.Context, p domain.Project) (domain.Project, error) {
	if err := p.Validate(); err != nil {
		return domain.Project{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	existing, ok := r.store.projects[p.ID]
	if !ok || existing.TenantID != p.TenantID {
		return domain.Project{}, domain.NewErrf(domain.CodeNotFound, "project %s not found", p.ID)
	}
	codeKey := p.TenantID + "|" + p.Code
	if id, ok := r.store.projectsByCode[codeKey]; ok && id != p.ID {
		return domain.Project{}, domain.NewErrf(domain.CodeAlreadyExists, "project code %s already exists", p.Code).WithField("code")
	}
	r.store.projects[p.ID] = p
	r.store.projectsByCode[codeKey] = p.ID
	return p, nil
}

func projectToFilterMap(p domain.Project) map[string]string {
	m := map[string]string{
		"code":       p.Code,
		"name":       p.Name,
		"sponsor":    p.Sponsor,
		"start_year": itoa(p.StartYear),
		"end_year":   itoa(p.EndYear),
	}
	for k, v := range p.Metadata {
		m["meta."+k] = v
	}
	return m
}

func sortProjects(in []domain.Project, orderBy string, desc bool) {
	switch orderBy {
	case "", "code":
		sort.Slice(in, func(i, j int) bool {
			if desc {
				return in[i].Code > in[j].Code
			}
			return in[i].Code < in[j].Code
		})
	case "name":
		sort.Slice(in, func(i, j int) bool {
			if desc {
				return in[i].Name > in[j].Name
			}
			return in[i].Name < in[j].Name
		})
	case "start_year":
		sort.Slice(in, func(i, j int) bool {
			if desc {
				return in[i].StartYear > in[j].StartYear
			}
			return in[i].StartYear < in[j].StartYear
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
