package memory

import (
	"context"
	"sort"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// RuleVersionRepo is the in-memory rule version repo.
type RuleVersionRepo struct {
	store *Store
}

func (r *RuleVersionRepo) Create(ctx context.Context, rv domain.RuleVersion) (domain.RuleVersion, error) {
	if err := rv.Validate(); err != nil {
		return domain.RuleVersion{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	codeKey := rv.TenantID + "|" + rv.Code
	if _, ok := r.store.rulesByCode[codeKey]; ok {
		return domain.RuleVersion{}, domain.NewErrf(domain.CodeAlreadyExists, "rule version code %s already exists", rv.Code).WithField("code")
	}
	r.store.rules[rv.ID] = rv
	r.store.rulesByCode[codeKey] = rv.ID
	return rv, nil
}

func (r *RuleVersionRepo) Get(ctx context.Context, tenantID, id string) (domain.RuleVersion, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	rv, ok := r.store.rules[id]
	if !ok || rv.TenantID != tenantID {
		return domain.RuleVersion{}, domain.NewErrf(domain.CodeNotFound, "rule version %s not found", id)
	}
	return rv, nil
}

func (r *RuleVersionRepo) GetByCode(ctx context.Context, tenantID, code string) (domain.RuleVersion, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	id, ok := r.store.rulesByCode[tenantID+"|"+code]
	if !ok {
		return domain.RuleVersion{}, domain.NewErrf(domain.CodeNotFound, "rule version code %s not found", code)
	}
	rv, _ := r.store.rules[id]
	return rv, nil
}

func (r *RuleVersionRepo) List(ctx context.Context, q application.ListQuery) ([]domain.RuleVersion, application.PageResult, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.RuleVersion, 0, len(r.store.rules))
	for _, rv := range r.store.rules {
		if rv.TenantID != q.TenantID {
			continue
		}
		if !matchesFilters(q.Filters, map[string]string{
			"code":       rv.Code,
			"project_id": rv.ProjectID,
			"status":     string(rv.Status),
		}) {
			continue
		}
		out = append(out, rv)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Code == out[j].Code {
			return out[i].ID < out[j].ID
		}
		return out[i].Code < out[j].Code
	})
	if q.OrderDesc {
		reverseRules(out)
	}
	page := pageSlice(out, q.Page, q.PageSize)
	res := application.PageResult{Page: q.Page, PageSize: q.PageSize, Total: len(out), HasNext: q.Page*q.PageSize < len(out)}
	return page, res, nil
}

func (r *RuleVersionRepo) Update(ctx context.Context, rv domain.RuleVersion) (domain.RuleVersion, error) {
	if err := rv.Validate(); err != nil {
		return domain.RuleVersion{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	existing, ok := r.store.rules[rv.ID]
	if !ok || existing.TenantID != rv.TenantID {
		return domain.RuleVersion{}, domain.NewErrf(domain.CodeNotFound, "rule version %s not found", rv.ID)
	}
	if existing.Version != rv.Version-1 {
		return domain.RuleVersion{}, domain.NewErrf(domain.CodeAborted, "stale rule version %s (expected %d, got %d)", rv.ID, existing.Version+1, rv.Version).WithField("version")
	}
	r.store.rules[rv.ID] = rv
	return rv, nil
}

func reverseRules(in []domain.RuleVersion) {
	for i, j := 0, len(in)-1; i < j; i, j = i+1, j-1 {
		in[i], in[j] = in[j], in[i]
	}
}
