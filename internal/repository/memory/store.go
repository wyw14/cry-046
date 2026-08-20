package memory

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/wyw14/cry-046/internal/domain"
)

var ErrNotFound = errors.New("entity not found")
var ErrConflict = errors.New("stale entity version")
var ErrDuplicate = errors.New("duplicate entity")

type Store struct {
	mu         sync.RWMutex
	projects   map[string]domain.Project
	assets     map[string]domain.Asset
	palettes   map[string]domain.Palette
	deliveries map[string]domain.DeliveryRequest
	audits     []domain.AuditEvent
}

func NewStore() *Store {
	return &Store{projects: map[string]domain.Project{}, assets: map[string]domain.Asset{}, palettes: map[string]domain.Palette{}, deliveries: map[string]domain.DeliveryRequest{}}
}
func cloneStrings(in []string) []string { return append([]string(nil), in...) }
func clonePalette(p domain.Palette) domain.Palette {
	p.Entries = append([]domain.ColorEntry(nil), p.Entries...)
	return p
}
func (s *Store) CreateProject(_ context.Context, p domain.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[p.ID]; ok {
		return ErrDuplicate
	}
	p.Tags = cloneStrings(p.Tags)
	s.projects[p.ID] = p
	return nil
}
func (s *Store) GetProject(_ context.Context, id string) (domain.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.projects[id]
	if !ok {
		return domain.Project{}, ErrNotFound
	}
	p.Tags = cloneStrings(p.Tags)
	return p, nil
}
func (s *Store) UpdateProject(_ context.Context, p domain.Project, expected int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.projects[p.ID]
	if !ok {
		return ErrNotFound
	}
	if old.Version != expected {
		return ErrConflict
	}
	p.Tags = cloneStrings(p.Tags)
	s.projects[p.ID] = p
	return nil
}
func (s *Store) ListProjects(_ context.Context, q string, page, size int) ([]domain.Project, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := make([]domain.Project, 0)
	for _, p := range s.projects {
		if q == "" || strings.Contains(strings.ToLower(p.Name+" "+p.Customer+" "+p.Series), strings.ToLower(q)) {
			p.Tags = cloneStrings(p.Tags)
			all = append(all, p)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.After(all[j].CreatedAt) })
	total := len(all)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}
func (s *Store) CreateAsset(_ context.Context, a domain.Asset) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.assets[a.ID]; ok {
		return ErrDuplicate
	}
	s.assets[a.ID] = a
	return nil
}
func (s *Store) GetAsset(_ context.Context, id string) (domain.Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.assets[id]
	if !ok {
		return domain.Asset{}, ErrNotFound
	}
	return a, nil
}
func (s *Store) ListAssets(_ context.Context, pid string) ([]domain.Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Asset{}
	for _, a := range s.assets {
		if a.ProjectID == pid {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
func (s *Store) UpdateAsset(_ context.Context, a domain.Asset, expected int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.assets[a.ID]
	if !ok {
		return ErrNotFound
	}
	if old.Version != expected {
		return ErrConflict
	}
	s.assets[a.ID] = a
	return nil
}
func (s *Store) CreatePalette(_ context.Context, p domain.Palette) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.palettes[p.ID]; ok {
		return ErrDuplicate
	}
	s.palettes[p.ID] = clonePalette(p)
	return nil
}
func (s *Store) GetPalette(_ context.Context, id string) (domain.Palette, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.palettes[id]
	if !ok {
		return domain.Palette{}, ErrNotFound
	}
	return clonePalette(p), nil
}
func (s *Store) UpdatePalette(_ context.Context, p domain.Palette, expected int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.palettes[p.ID]
	if !ok {
		return ErrNotFound
	}
	if old.Revision != expected {
		return ErrConflict
	}
	s.palettes[p.ID] = clonePalette(p)
	return nil
}
func (s *Store) ListPalettes(_ context.Context, pid string) ([]domain.Palette, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Palette{}
	for _, p := range s.palettes {
		if p.ProjectID == pid {
			out = append(out, clonePalette(p))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
func (s *Store) CreateDelivery(_ context.Context, d domain.DeliveryRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.deliveries[d.ID]; ok {
		return ErrDuplicate
	}
	s.deliveries[d.ID] = d
	return nil
}
func (s *Store) GetDelivery(_ context.Context, id string) (domain.DeliveryRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.deliveries[id]
	if !ok {
		return domain.DeliveryRequest{}, ErrNotFound
	}
	return d, nil
}
func (s *Store) UpdateDelivery(_ context.Context, d domain.DeliveryRequest, expected int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.deliveries[d.ID]
	if !ok {
		return ErrNotFound
	}
	if old.Revision != expected {
		return ErrConflict
	}
	s.deliveries[d.ID] = d
	return nil
}
func (s *Store) ListDeliveries(_ context.Context, pid string) ([]domain.DeliveryRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.DeliveryRequest{}
	for _, d := range s.deliveries {
		p, ok := s.palettes[d.PaletteID]
		if ok && p.ProjectID == pid {
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (s *Store) Append(_ context.Context, e domain.AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audits = append(s.audits, e)
	return nil
}
func (s *Store) List(_ context.Context, entity string, limit int) ([]domain.AuditEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.AuditEvent{}
	for i := len(s.audits) - 1; i >= 0 && len(out) < limit; i-- {
		if entity == "" || s.audits[i].EntityID == entity {
			out = append(out, s.audits[i])
		}
	}
	return out, nil
}
