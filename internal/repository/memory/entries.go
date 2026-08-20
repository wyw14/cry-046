package memory

import (
	"context"
	"sort"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// EntryRepo is the in-memory settlement entry repo.
type EntryRepo struct {
	store *Store
}

// UpsertBatch inserts new entries and updates existing ones, keyed by
// the deduplication fingerprint. The returned summary reports the
// number of created, updated and skipped rows.
func (r *EntryRepo) UpsertBatch(ctx context.Context, entries []domain.SettlementEntry) (application.UpsertSummary, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	var s application.UpsertSummary
	for _, e := range entries {
		if err := e.Validate(); err != nil {
			return s, err
		}
		if id, ok := r.store.entryByFP[e.SourceFingerprint]; ok {
			existing := r.store.entries[id]
			if existing.TenantID != e.TenantID {
				return s, domain.NewErrf(domain.CodeConflict, "fingerprint collision across tenants").WithField("source_fingerprint")
			}
			e.ID = existing.ID
			e.CreatedAt = existing.CreatedAt
			r.store.entries[id] = e
			s.Updated++
			continue
		}
		r.store.entries[e.ID] = e
		r.store.entryByFP[e.SourceFingerprint] = e.ID
		s.Created++
	}
	return s, nil
}

func (r *EntryRepo) Get(ctx context.Context, tenantID, id string) (domain.SettlementEntry, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	e, ok := r.store.entries[id]
	if !ok || e.TenantID != tenantID {
		return domain.SettlementEntry{}, domain.NewErrf(domain.CodeNotFound, "entry %s not found", id)
	}
	return e, nil
}

func (r *EntryRepo) List(ctx context.Context, q application.EntryListQuery) ([]domain.SettlementEntry, application.PageResult, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.SettlementEntry, 0, len(r.store.entries))
	for _, e := range r.store.entries {
		if e.TenantID != q.TenantID {
			continue
		}
		if q.CycleID != "" && e.CycleID != q.CycleID {
			continue
		}
		if q.BatchID != "" && e.BatchID != q.BatchID {
			continue
		}
		if q.ProjectID != "" && e.ProjectID != q.ProjectID {
			continue
		}
		if q.Source != "" && e.Source != domain.EntrySource(q.Source) {
			continue
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].OccurredAt.Equal(out[j].OccurredAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].OccurredAt.Before(out[j].OccurredAt)
	})
	if q.OrderDesc {
		reverseEntries(out)
	}
	page := pageSlice(out, q.Page, q.PageSize)
	res := application.PageResult{Page: q.Page, PageSize: q.PageSize, Total: len(out), HasNext: q.Page*q.PageSize < len(out)}
	return page, res, nil
}

func (r *EntryRepo) ListByCycle(ctx context.Context, tenantID, cycleID string) ([]domain.SettlementEntry, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.SettlementEntry, 0)
	for _, e := range r.store.entries {
		if e.TenantID != tenantID || e.CycleID != cycleID {
			continue
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].OccurredAt.Equal(out[j].OccurredAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].OccurredAt.Before(out[j].OccurredAt)
	})
	return out, nil
}

func reverseEntries(in []domain.SettlementEntry) {
	for i, j := 0, len(in)-1; i < j; i, j = i+1, j-1 {
		in[i], in[j] = in[j], in[i]
	}
}

// ExceptionRepo is the in-memory exception repo.
type ExceptionRepo struct {
	store *Store
}

func (r *ExceptionRepo) Create(ctx context.Context, e domain.Exception) (domain.Exception, error) {
	if err := e.Validate(); err != nil {
		return domain.Exception{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, ok := r.store.exceptions[e.ID]; ok {
		return domain.Exception{}, domain.NewErrf(domain.CodeAlreadyExists, "exception %s already exists", e.ID)
	}
	r.store.exceptions[e.ID] = e
	return e, nil
}

func (r *ExceptionRepo) Get(ctx context.Context, tenantID, id string) (domain.Exception, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	e, ok := r.store.exceptions[id]
	if !ok || e.TenantID != tenantID {
		return domain.Exception{}, domain.NewErrf(domain.CodeNotFound, "exception %s not found", id)
	}
	return e, nil
}

func (r *ExceptionRepo) Update(ctx context.Context, e domain.Exception) (domain.Exception, error) {
	if err := e.Validate(); err != nil {
		return domain.Exception{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	existing, ok := r.store.exceptions[e.ID]
	if !ok || existing.TenantID != e.TenantID {
		return domain.Exception{}, domain.NewErrf(domain.CodeNotFound, "exception %s not found", e.ID)
	}
	if e.Version != existing.Version+1 {
		return domain.Exception{}, domain.NewErrf(domain.CodeAborted, "stale exception %s (expected %d, got %d)", e.ID, existing.Version+1, e.Version).WithField("version")
	}
	r.store.exceptions[e.ID] = e
	return e, nil
}

func (r *ExceptionRepo) List(ctx context.Context, q application.ExceptionListQuery) ([]domain.Exception, application.PageResult, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.Exception, 0, len(r.store.exceptions))
	for _, e := range r.store.exceptions {
		if e.TenantID != q.TenantID {
			continue
		}
		if q.CycleID != "" && e.CycleID != q.CycleID {
			continue
		}
		if q.EntryID != "" && e.EntryID != q.EntryID {
			continue
		}
		if q.Status != "" && e.Status != domain.ExceptionStatus(q.Status) {
			continue
		}
		if q.Severity != "" && e.Severity != domain.Severity(q.Severity) {
			continue
		}
		if q.AssigneeID != "" && e.AssigneeID != q.AssigneeID {
			continue
		}
		if q.OverdueOnly && !e.IsOverdue(q.AsOf) {
			continue
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Severity.Weight() != out[j].Severity.Weight() {
			return out[i].Severity.Weight() > out[j].Severity.Weight()
		}
		if !out[i].DeadlineAt.Equal(out[j].DeadlineAt) {
			if out[i].DeadlineAt.IsZero() {
				return false
			}
			if out[j].DeadlineAt.IsZero() {
				return true
			}
			return out[i].DeadlineAt.Before(out[j].DeadlineAt)
		}
		return out[i].ID < out[j].ID
	})
	if q.OrderDesc {
		reverseExceptions(out)
	}
	page := pageSlice(out, q.Page, q.PageSize)
	res := application.PageResult{Page: q.Page, PageSize: q.PageSize, Total: len(out), HasNext: q.Page*q.PageSize < len(out)}
	return page, res, nil
}

func (r *ExceptionRepo) ListByAssignee(ctx context.Context, tenantID, assigneeID string) ([]domain.Exception, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.Exception, 0)
	for _, e := range r.store.exceptions {
		if e.TenantID != tenantID {
			continue
		}
		if e.AssigneeID != assigneeID {
			continue
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *ExceptionRepo) ListByCycle(ctx context.Context, tenantID, cycleID string) ([]domain.Exception, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.Exception, 0)
	for _, e := range r.store.exceptions {
		if e.TenantID != tenantID || e.CycleID != cycleID {
			continue
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func reverseExceptions(in []domain.Exception) {
	for i, j := 0, len(in)-1; i < j; i, j = i+1, j-1 {
		in[i], in[j] = in[j], in[i]
	}
}
