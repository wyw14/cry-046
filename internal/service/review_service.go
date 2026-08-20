package service

import (
	"context"
	"errors"
	"time"

	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/domain"
)

type ReviewService struct {
	palettes application.PaletteRepository
	audits   application.AuditRepository
	clock    application.Clock
	ids      application.IDGenerator
}

func NewReviewService(p application.PaletteRepository, a application.AuditRepository, c application.Clock, i application.IDGenerator) *ReviewService {
	return &ReviewService{palettes: p, audits: a, clock: c, ids: i}
}
func (s *ReviewService) Withdraw(ctx context.Context, id, actor string) (domain.Palette, error) {
	p, err := s.palettes.GetPalette(ctx, id)
	if err != nil {
		return domain.Palette{}, err
	}
	if p.Status == domain.StatusDelivered {
		return domain.Palette{}, errors.New("delivered palette cannot be withdrawn")
	}
	if err := domain.TransitionProposal(p.Status, domain.StatusWithdrawn); err != nil {
		return domain.Palette{}, err
	}
	p.Status = domain.StatusWithdrawn
	p.Revision++
	if err := s.palettes.UpdatePalette(ctx, p, p.Revision-1); err != nil {
		return domain.Palette{}, err
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor, Action: "palette.withdrawn", Entity: "palette", EntityID: id, CreatedAt: s.clock.Now()})
	return p, nil
}
func (s *ReviewService) Archive(ctx context.Context, id, actor string) (domain.Palette, error) {
	p, err := s.palettes.GetPalette(ctx, id)
	if err != nil {
		return domain.Palette{}, err
	}
	if p.Status != domain.StatusWithdrawn {
		return domain.Palette{}, errors.New("only withdrawn palette can be archived")
	}
	p.Status = domain.StatusArchived
	p.Revision++
	if err := s.palettes.UpdatePalette(ctx, p, p.Revision-1); err != nil {
		return domain.Palette{}, err
	}
	return p, nil
}
func (s *ReviewService) Comment(ctx context.Context, id string, comment domain.ReviewComment) error {
	p, err := s.palettes.GetPalette(ctx, id)
	if err != nil {
		return err
	}
	if p.Status == domain.StatusArchived {
		return errors.New("archived palette cannot receive comments")
	}
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = time.Now().UTC()
	}
	_ = p
	return nil
}
