package memory

import (
	"context"
	"sort"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// AuditRepo is the in-memory audit log repo.
type AuditRepo struct {
	store *Store
}

func (r *AuditRepo) Append(ctx context.Context, e domain.AuditEntry) (domain.AuditEntry, error) {
	if err := e.Validate(); err != nil {
		return domain.AuditEntry{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.audits = append(r.store.audits, e)
	return e, nil
}

func (r *AuditRepo) List(ctx context.Context, q application.ListQuery) ([]domain.AuditEntry, application.PageResult, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.AuditEntry, 0, len(r.store.audits))
	for _, e := range r.store.audits {
		if e.TenantID != q.TenantID {
			continue
		}
		if !matchesFilters(q.Filters, map[string]string{
			"actor_id":    e.ActorID,
			"action":      string(e.Action),
			"entity_id":   e.EntityID,
			"entity_type": e.EntityType,
		}) {
			continue
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	page := pageSlice(out, q.Page, q.PageSize)
	res := application.PageResult{Page: q.Page, PageSize: q.PageSize, Total: len(out), HasNext: q.Page*q.PageSize < len(out)}
	return page, res, nil
}

// UserRepo is the in-memory user repo.
type UserRepo struct {
	store *Store
}

func (r *UserRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	if err := u.Validate(); err != nil {
		return domain.User{}, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	key := u.TenantID + "|" + u.Username
	if _, ok := r.store.usersByName[key]; ok {
		return domain.User{}, domain.NewErrf(domain.CodeAlreadyExists, "username %s already exists", u.Username).WithField("username")
	}
	r.store.users[u.ID] = u
	r.store.usersByName[key] = u.ID
	return u, nil
}

func (r *UserRepo) Get(ctx context.Context, tenantID, id string) (domain.User, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	u, ok := r.store.users[id]
	if !ok || u.TenantID != tenantID {
		return domain.User{}, domain.NewErrf(domain.CodeNotFound, "user %s not found", id)
	}
	return u, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, tenantID, username string) (domain.User, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	id, ok := r.store.usersByName[tenantID+"|"+username]
	if !ok {
		return domain.User{}, domain.NewErrf(domain.CodeNotFound, "user %s not found", username)
	}
	u, _ := r.store.users[id]
	return u, nil
}

func (r *UserRepo) List(ctx context.Context, q application.ListQuery) ([]domain.User, application.PageResult, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]domain.User, 0, len(r.store.users))
	for _, u := range r.store.users {
		if u.TenantID != q.TenantID {
			continue
		}
		if !matchesFilters(q.Filters, map[string]string{
			"username": u.Username,
			"role":     string(u.Role),
		}) {
			continue
		}
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Username < out[j].Username })
	page := pageSlice(out, q.Page, q.PageSize)
	res := application.PageResult{Page: q.Page, PageSize: q.PageSize, Total: len(out), HasNext: q.Page*q.PageSize < len(out)}
	return page, res, nil
}
