package service

import (
	"context"
	"errors"
	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/domain"
	"sort"
	"sync"
	"time"
)

type Workspace struct {
	mu        sync.RWMutex
	projects  application.ProjectRepository
	favorites map[string]domain.Favorite
	recent    map[string][]domain.RecentVisit
	todos     map[string]domain.Todo
	clock     application.Clock
	ids       application.IDGenerator
}

func NewWorkspace(p application.ProjectRepository, c application.Clock, i application.IDGenerator) *Workspace {
	return &Workspace{projects: p, favorites: map[string]domain.Favorite{}, recent: map[string][]domain.RecentVisit{}, todos: map[string]domain.Todo{}, clock: c, ids: i}
}
func favKey(user, project string) string { return user + "\x00" + project }
func (w *Workspace) ToggleFavorite(ctx context.Context, user, project string) (bool, error) {
	if _, err := w.projects.GetProject(ctx, project); err != nil {
		return false, err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	k := favKey(user, project)
	if _, ok := w.favorites[k]; ok {
		delete(w.favorites, k)
		return false, nil
	}
	w.favorites[k] = domain.Favorite{UserID: user, ProjectID: project, CreatedAt: w.clock.Now()}
	return true, nil
}
func (w *Workspace) RecordVisit(ctx context.Context, user, project string) error {
	if _, err := w.projects.GetProject(ctx, project); err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	list := w.recent[user]
	out := make([]domain.RecentVisit, 0, len(list)+1)
	for _, v := range list {
		if v.ProjectID != project {
			out = append(out, v)
		}
	}
	out = append([]domain.RecentVisit{{UserID: user, ProjectID: project, VisitedAt: w.clock.Now()}}, out...)
	if len(out) > 50 {
		out = out[:50]
	}
	w.recent[user] = out
	return nil
}
func (w *Workspace) ListFavorites(_ context.Context, user string) ([]domain.Favorite, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := []domain.Favorite{}
	for _, f := range w.favorites {
		if f.UserID == user {
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (w *Workspace) ListRecent(_ context.Context, user string, limit int) ([]domain.RecentVisit, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := append([]domain.RecentVisit(nil), w.recent[user]...)
	if limit < 1 || limit > len(out) {
		limit = len(out)
	}
	return out[:limit], nil
}
func (w *Workspace) AddTodo(user, project, kind, summary string, due time.Time) domain.Todo {
	w.mu.Lock()
	defer w.mu.Unlock()
	t := domain.Todo{ID: w.ids.NewID("todo"), UserID: user, ProjectID: project, Kind: kind, Summary: summary, DueAt: due, CreatedAt: w.clock.Now()}
	w.todos[t.ID] = t
	return t
}
func (w *Workspace) ListTodos(_ context.Context, user string, includeDone bool) ([]domain.Todo, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := []domain.Todo{}
	for _, t := range w.todos {
		if t.UserID == user && (includeDone || !t.Done) {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DueAt.Before(out[j].DueAt) })
	return out, nil
}
func (w *Workspace) CompleteTodo(_ context.Context, user, id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	t, ok := w.todos[id]
	if !ok {
		return errors.New("todo not found")
	}
	if t.UserID != user {
		return ErrForbidden
	}
	t.Done = true
	w.todos[id] = t
	return nil
}

var _ application.WorkspaceService = (*Workspace)(nil)
