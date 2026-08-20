package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/domain"
)

type PaletteService struct {
	palettes application.PaletteRepository
	projects application.ProjectRepository
	assets   application.AssetRepository
	audits   application.AuditRepository
	clock    application.Clock
	ids      application.IDGenerator
}

func NewPaletteService(p application.PaletteRepository, pr application.ProjectRepository, a application.AssetRepository, au application.AuditRepository, c application.Clock, i application.IDGenerator) *PaletteService {
	return &PaletteService{palettes: p, projects: pr, assets: a, audits: au, clock: c, ids: i}
}
func (s *PaletteService) Create(ctx context.Context, in application.CreatePaletteInput) (domain.Palette, error) {
	if _, err := s.projects.GetProject(ctx, in.ProjectID); err != nil {
		return domain.Palette{}, err
	}
	p := domain.Palette{ID: in.ID, ProjectID: in.ProjectID, Name: strings.TrimSpace(in.Name), Source: in.Source, Branch: in.Branch, Version: 1, Revision: 1, Status: domain.StatusDraft, CreatedAt: s.clock.Now()}
	for _, c := range in.Entries {
		p.Entries = append(p.Entries, domain.ColorEntry{Name: c.Name, Hex: c.Hex, Source: c.Source, Replacement: c.Replacement, AssetID: c.AssetID})
	}
	if err := p.Validate(); err != nil {
		return domain.Palette{}, err
	}
	if err := s.validateAssets(ctx, p); err != nil {
		return domain.Palette{}, err
	}
	if err := s.palettes.CreatePalette(ctx, p); err != nil {
		return domain.Palette{}, err
	}
	s.audit(ctx, "palette.created", p.ID)
	return p, nil
}
func (s *PaletteService) validateAssets(ctx context.Context, p domain.Palette) error {
	for _, e := range p.Entries {
		if e.AssetID == "" {
			continue
		}
		if _, err := s.assets.GetAsset(ctx, e.AssetID); err != nil {
			return fmt.Errorf("asset %s: %w", e.AssetID, err)
		}
	}
	return nil
}
func (s *PaletteService) Get(ctx context.Context, id string) (domain.Palette, error) {
	return s.palettes.GetPalette(ctx, id)
}
func (s *PaletteService) List(ctx context.Context, pid string) ([]domain.Palette, error) {
	return s.palettes.ListPalettes(ctx, pid)
}
func (s *PaletteService) Submit(ctx context.Context, id, actor string) (domain.Palette, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return domain.Palette{}, err
	}
	if err := domain.TransitionProposal(p.Status, domain.StatusReview); err != nil {
		return domain.Palette{}, err
	}
	p.Status = domain.StatusReview
	p.Revision++
	if err := s.palettes.UpdatePalette(ctx, p, p.Revision-1); err != nil {
		return domain.Palette{}, err
	}
	s.audit(ctx, "palette.submitted", id)
	return p, nil
}
func (s *PaletteService) Approve(ctx context.Context, id, actor string) (domain.Palette, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return domain.Palette{}, err
	}
	if strings.TrimSpace(actor) == "" {
		return domain.Palette{}, errors.New("reviewer required")
	}
	if err := domain.TransitionProposal(p.Status, domain.StatusApproved); err != nil {
		return domain.Palette{}, err
	}
	p.Status = domain.StatusApproved
	p.Revision++
	if err := s.palettes.UpdatePalette(ctx, p, p.Revision-1); err != nil {
		return domain.Palette{}, err
	}
	s.audit(ctx, "palette.approved", id)
	return p, nil
}
func (s *PaletteService) Deliver(ctx context.Context, id, actor string) (domain.Palette, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return domain.Palette{}, err
	}
	if err := domain.TransitionProposal(p.Status, domain.StatusDelivered); err != nil {
		return domain.Palette{}, err
	}
	p.Status = domain.StatusDelivered
	now := s.clock.Now()
	p.DeliveredAt = &now
	p.Revision++
	if err := s.palettes.UpdatePalette(ctx, p, p.Revision-1); err != nil {
		return domain.Palette{}, err
	}
	s.audit(ctx, "palette.delivered", id)
	return p, nil
}
func (s *PaletteService) Derive(ctx context.Context, id, newID, name, actor string) (domain.Palette, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return domain.Palette{}, err
	}
	d, err := p.Derive(newID, name)
	if err != nil {
		return domain.Palette{}, err
	}
	if err := s.palettes.CreatePalette(ctx, d); err != nil {
		return domain.Palette{}, err
	}
	s.audit(ctx, "palette.derived", d.ID)
	return d, nil
}
func (s *PaletteService) Diff(ctx context.Context, left, right string) (map[string]any, error) {
	a, err := s.Get(ctx, left)
	if err != nil {
		return nil, err
	}
	b, err := s.Get(ctx, right)
	if err != nil {
		return nil, err
	}
	changes := map[string]any{"added": []domain.ColorEntry{}, "removed": []domain.ColorEntry{}, "changed": []map[string]string{}}
	bm := map[string]domain.ColorEntry{}
	for _, e := range b.Entries {
		bm[e.Name] = e
	}
	am := map[string]domain.ColorEntry{}
	for _, e := range a.Entries {
		am[e.Name] = e
	}
	added, removed := []domain.ColorEntry{}, []domain.ColorEntry{}
	changed := []map[string]string{}
	for n, e := range bm {
		if old, ok := am[n]; !ok {
			added = append(added, e)
		} else if old.Hex != e.Hex {
			changed = append(changed, map[string]string{"name": n, "from": old.Hex, "to": e.Hex})
		}
	}
	for n, e := range am {
		if _, ok := bm[n]; !ok {
			removed = append(removed, e)
		}
	}
	changes["added"], changes["removed"], changes["changed"] = added, removed, changed
	return changes, nil
}
func (s *PaletteService) audit(ctx context.Context, action, id string) {
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: "system", Action: action, Entity: "palette", EntityID: id, CreatedAt: s.clock.Now()})
}
