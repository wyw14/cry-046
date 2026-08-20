package memory

import (
	"context"
	"sort"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// PartyRepo is the in-memory PartyRepo.
type PartyRepo struct {
	store *Store
}

func (r *PartyRepo) Create(ctx context.Context, p domain.Party) (domain.Party, error) {
	if err := p.Validate(); err != nil {
		return domain.Party{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.parties[p.ID] = p
	return p, nil
}

func (r *PartyRepo) Get(ctx context.Context, tenantID, id string) (domain.Party, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	p, ok := r.store.parties[id]
	if !ok || p.TenantID != tenantID {
		return domain.Party{}, domain.NewErrf(domain.CodeNotFound, "party %s not found", id)
	}
	return p, nil
}

func (r *PartyRepo) List(ctx context.Context, q application.ListQuery) ([]domain.Party, application.PageResult, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.Party, 0, len(r.store.parties))
	for _, p := range r.store.parties {
		if p.TenantID != q.TenantID {
			continue
		}
		if !matchesFilters(q.Filters, partyToFilterMap(p)) {
			continue
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name == out[j].Name {
			return out[i].ID < out[j].ID
		}
		return out[i].Name < out[j].Name
	})
	if q.OrderDesc {
		reverseParties(out)
	}
	page := pageSlice(out, q.Page, q.PageSize)
	res := application.PageResult{Page: q.Page, PageSize: q.PageSize, Total: len(out), HasNext: q.Page*q.PageSize < len(out)}
	return page, res, nil
}

func partyToFilterMap(p domain.Party) map[string]string {
	m := map[string]string{
		"type": string(p.Type),
		"name": p.Name,
	}
	for k, v := range p.Metadata {
		m["meta."+k] = v
	}
	return m
}

func reverseParties(in []domain.Party) {
	for i, j := 0, len(in)-1; i < j; i, j = i+1, j-1 {
		in[i], in[j] = in[j], in[i]
	}
}
