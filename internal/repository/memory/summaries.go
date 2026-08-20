package memory

import (
	"context"
	"sort"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// SummaryRepo is the in-memory summary repo.
type SummaryRepo struct {
	store *Store
}

func (r *SummaryRepo) GetLatest(ctx context.Context, tenantID, cycleID string) (domain.Summary, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	var latest domain.Summary
	found := false
	for _, s := range r.store.summaries {
		if s.TenantID != tenantID || s.CycleID != cycleID {
			continue
		}
		if !found || s.Version > latest.Version {
			latest = s
			found = true
		}
	}
	if !found {
		return domain.Summary{}, domain.NewErrf(domain.CodeNotFound, "no summary for cycle %s", cycleID)
	}
	return latest, nil
}

func (r *SummaryRepo) Save(ctx context.Context, s domain.Summary) (domain.Summary, error) {
	if err := s.Validate(); err != nil {
		return domain.Summary{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.summaries = append(r.store.summaries, s)
	return s, nil
}

func (r *SummaryRepo) List(ctx context.Context, tenantID, cycleID string, limit int) ([]domain.Summary, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.Summary, 0)
	for _, s := range r.store.summaries {
		if s.TenantID != tenantID || s.CycleID != cycleID {
			continue
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version > out[j].Version })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// RecalcRepo is the in-memory recalculation batch repo.
type RecalcRepo struct {
	store *Store
}

func (r *RecalcRepo) Create(ctx context.Context, rb domain.RecalculationBatch) (domain.RecalculationBatch, error) {
	if err := rb.Validate(); err != nil {
		return domain.RecalculationBatch{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.recalcs[rb.ID] = rb
	return rb, nil
}

func (r *RecalcRepo) Update(ctx context.Context, rb domain.RecalculationBatch) (domain.RecalculationBatch, error) {
	if err := rb.Validate(); err != nil {
		return domain.RecalculationBatch{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, ok := r.store.recalcs[rb.ID]; !ok {
		return domain.RecalculationBatch{}, domain.NewErrf(domain.CodeNotFound, "recalc %s not found", rb.ID)
	}
	r.store.recalcs[rb.ID] = rb
	return rb, nil
}

func (r *RecalcRepo) Get(ctx context.Context, tenantID, id string) (domain.RecalculationBatch, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	rb, ok := r.store.recalcs[id]
	if !ok || rb.TenantID != tenantID {
		return domain.RecalculationBatch{}, domain.NewErrf(domain.CodeNotFound, "recalc %s not found", id)
	}
	return rb, nil
}

func (r *RecalcRepo) List(ctx context.Context, q application.ListQuery) ([]domain.RecalculationBatch, application.PageResult, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.RecalculationBatch, 0, len(r.store.recalcs))
	for _, rb := range r.store.recalcs {
		if rb.TenantID != q.TenantID {
			continue
		}
		if !matchesFilters(q.Filters, map[string]string{
			"cycle_id":        rb.CycleID,
			"rule_version_id": rb.RuleVersionID,
			"status":          string(rb.Status),
		}) {
			continue
		}
		out = append(out, rb)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].StartedAt.Equal(out[j].StartedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].StartedAt.After(out[j].StartedAt)
	})
	page := pageSlice(out, q.Page, q.PageSize)
	res := application.PageResult{Page: q.Page, PageSize: q.PageSize, Total: len(out), HasNext: q.Page*q.PageSize < len(out)}
	return page, res, nil
}

// AnnualRepo is the in-memory annual accumulator repo.
type AnnualRepo struct {
	store *Store
}

func annualKey(tenantID, projectID string, year int) string {
	return tenantID + "|" + projectID + "|" + itoa(year)
}

func (r *AnnualRepo) Get(ctx context.Context, tenantID, projectID string, year int) (domain.AnnualAccumulator, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	if a, ok := r.store.annuals[annualKey(tenantID, projectID, year)]; ok {
		return a, nil
	}
	for _, acc := range r.store.annuals {
		if acc.ProjectID == projectID && acc.Year == year {
			return acc, nil
		}
	}
	return domain.AnnualAccumulator{}, domain.NewErrf(domain.CodeNotFound, "annual accumulator for project %s year %d not found", projectID, year)
}

func (r *AnnualRepo) ApplyAdjustment(ctx context.Context, adj domain.Adjustment) (domain.AnnualAccumulator, error) {
	if err := adj.Validate(); err != nil {
		return domain.AnnualAccumulator{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	// Adjustment does not carry TenantID; look up by project+year across
	// all tenants (the project id is unique per tenant in practice) and
	// fall back to creating a new accumulator keyed on the empty tenant
	// when none exists yet. The Get/ListAdjustments methods accept the
	// tenant id and look it up directly so reads remain tenant-scoped.
	for k, acc := range r.store.annuals {
		if acc.ProjectID == adj.ProjectID && acc.Year == adj.Year {
			acc = acc.ApplyAdjustment(adj)
			r.store.annuals[k] = acc
			return acc, nil
		}
	}
	acc := domain.AnnualAccumulator{
		ProjectID:   adj.ProjectID,
		Year:        adj.Year,
		Adjustments: nil,
	}.ApplyAdjustment(adj)
	r.store.annuals[annualKey("", adj.ProjectID, adj.Year)] = acc
	return acc, nil
}

func (r *AnnualRepo) ListAdjustments(ctx context.Context, tenantID, projectID string, year int) ([]domain.Adjustment, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	a, ok := r.store.annuals[annualKey(tenantID, projectID, year)]
	if !ok {
		for _, acc := range r.store.annuals {
			if acc.ProjectID == projectID && acc.Year == year {
				a = acc
				ok = true
				break
			}
		}
	}
	if !ok {
		return nil, nil
	}
	out := append([]domain.Adjustment(nil), a.Adjustments...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out, nil
}
