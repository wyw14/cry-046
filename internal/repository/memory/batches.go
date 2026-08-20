package memory

import (
	"context"
	"sort"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// BatchRepo is the in-memory FundingBatch repo.
type BatchRepo struct {
	store *Store
}

func (r *BatchRepo) Create(ctx context.Context, b domain.FundingBatch) (domain.FundingBatch, error) {
	if err := b.Validate(); err != nil {
		return domain.FundingBatch{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	codeKey := b.TenantID + "|" + b.Code
	if _, ok := r.store.batchesByCode[codeKey]; ok {
		return domain.FundingBatch{}, domain.NewErrf(domain.CodeAlreadyExists, "batch code %s already exists", b.Code).WithField("code")
	}
	r.store.batches[b.ID] = b
	r.store.batchesByCode[codeKey] = b.ID
	return b, nil
}

func (r *BatchRepo) Get(ctx context.Context, tenantID, id string) (domain.FundingBatch, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	b, ok := r.store.batches[id]
	if !ok || b.TenantID != tenantID {
		return domain.FundingBatch{}, domain.NewErrf(domain.CodeNotFound, "batch %s not found", id)
	}
	return b, nil
}

func (r *BatchRepo) List(ctx context.Context, q application.ListQuery) ([]domain.FundingBatch, application.PageResult, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.FundingBatch, 0, len(r.store.batches))
	for _, b := range r.store.batches {
		if b.TenantID != q.TenantID {
			continue
		}
		if !matchesFilters(q.Filters, map[string]string{
			"code":       b.Code,
			"project_id": b.ProjectID,
		}) {
			continue
		}
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Code == out[j].Code {
			return out[i].ID < out[j].ID
		}
		return out[i].Code < out[j].Code
	})
	if q.OrderDesc {
		reverseBatches(out)
	}
	page := pageSlice(out, q.Page, q.PageSize)
	res := application.PageResult{Page: q.Page, PageSize: q.PageSize, Total: len(out), HasNext: q.Page*q.PageSize < len(out)}
	return page, res, nil
}

func reverseBatches(in []domain.FundingBatch) {
	for i, j := 0, len(in)-1; i < j; i, j = i+1, j-1 {
		in[i], in[j] = in[j], in[i]
	}
}

// CycleRepo is the in-memory SettlementCycle repo.
type CycleRepo struct {
	store *Store
}

func (r *CycleRepo) Create(ctx context.Context, c domain.SettlementCycle) (domain.SettlementCycle, error) {
	if err := c.Validate(); err != nil {
		return domain.SettlementCycle{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.cycles[c.ID] = c
	return c, nil
}

func (r *CycleRepo) Get(ctx context.Context, tenantID, id string) (domain.SettlementCycle, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	c, ok := r.store.cycles[id]
	if !ok || c.TenantID != tenantID {
		return domain.SettlementCycle{}, domain.NewErrf(domain.CodeNotFound, "cycle %s not found", id)
	}
	return c, nil
}

func (r *CycleRepo) List(ctx context.Context, q application.ListQuery) ([]domain.SettlementCycle, application.PageResult, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.SettlementCycle, 0, len(r.store.cycles))
	for _, c := range r.store.cycles {
		if c.TenantID != q.TenantID {
			continue
		}
		if !matchesFilters(q.Filters, map[string]string{
			"project_id":       c.ProjectID,
			"funding_batch_id": c.FundingBatchID,
			"year":             itoa(c.Year),
			"quarter":          itoa(c.Quarter),
		}) {
			continue
		}
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Year != out[j].Year {
			return out[i].Year > out[j].Year
		}
		if out[i].Quarter != out[j].Quarter {
			return out[i].Quarter > out[j].Quarter
		}
		return out[i].ID < out[j].ID
	})
	if q.OrderDesc {
		reverseCycles(out)
	}
	page := pageSlice(out, q.Page, q.PageSize)
	res := application.PageResult{Page: q.Page, PageSize: q.PageSize, Total: len(out), HasNext: q.Page*q.PageSize < len(out)}
	return page, res, nil
}

func (r *CycleRepo) Update(ctx context.Context, c domain.SettlementCycle) (domain.SettlementCycle, error) {
	if err := c.Validate(); err != nil {
		return domain.SettlementCycle{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	existing, ok := r.store.cycles[c.ID]
	if !ok || existing.TenantID != c.TenantID {
		return domain.SettlementCycle{}, domain.NewErrf(domain.CodeNotFound, "cycle %s not found", c.ID)
	}
	r.store.cycles[c.ID] = c
	return c, nil
}

func reverseCycles(in []domain.SettlementCycle) {
	for i, j := 0, len(in)-1; i < j; i, j = i+1, j-1 {
		in[i], in[j] = in[j], in[i]
	}
}
