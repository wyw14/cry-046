package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/domain"
)

type DeliveryService struct {
	deliveries application.DeliveryRepository
	palettes   application.PaletteRepository
	assets     application.AssetRepository
	projects   application.ProjectRepository
	writer     application.DeliveryWriter
	audits     application.AuditRepository
	notifier   application.Notifier
	clock      application.Clock
	ids        application.IDGenerator
}

func NewDeliveryService(d application.DeliveryRepository, p application.PaletteRepository, a application.AssetRepository, pr application.ProjectRepository, w application.DeliveryWriter, au application.AuditRepository, n application.Notifier, c application.Clock, i application.IDGenerator) *DeliveryService {
	return &DeliveryService{deliveries: d, palettes: p, assets: a, projects: pr, writer: w, audits: au, notifier: n, clock: c, ids: i}
}
func (s *DeliveryService) Request(ctx context.Context, in application.CreateDeliveryInput) (domain.DeliveryRequest, error) {
	p, err := s.palettes.GetPalette(ctx, in.PaletteID)
	if err != nil {
		return domain.DeliveryRequest{}, err
	}
	if p.Status != domain.StatusApproved && p.Status != domain.StatusDelivered {
		return domain.DeliveryRequest{}, errors.New("palette must be approved before delivery")
	}
	exp, err := time.Parse(time.RFC3339, in.ExpiresAt)
	if err != nil {
		return domain.DeliveryRequest{}, err
	}
	if !exp.After(s.clock.Now()) {
		return domain.DeliveryRequest{}, errors.New("expiry must be in the future")
	}
	d := domain.DeliveryRequest{ID: in.ID, PaletteID: p.ID, Applicant: in.Applicant, Format: strings.ToLower(in.Format), Status: domain.DeliveryRequested, ExpiresAt: exp, CreatedAt: s.clock.Now(), Revision: 1}
	if d.Format != "zip" && d.Format != "json" && d.Format != "csv" {
		return domain.DeliveryRequest{}, errors.New("unsupported delivery format")
	}
	if err := s.deliveries.CreateDelivery(ctx, d); err != nil {
		return domain.DeliveryRequest{}, err
	}
	s.audit(ctx, "delivery.requested", d.ID)
	return d, nil
}
func (s *DeliveryService) Approve(ctx context.Context, id, actor string) (domain.DeliveryRequest, error) {
	d, err := s.deliveries.GetDelivery(ctx, id)
	if err != nil {
		return domain.DeliveryRequest{}, err
	}
	p, err := s.palettes.GetPalette(ctx, d.PaletteID)
	if err != nil {
		return domain.DeliveryRequest{}, err
	}
	pr, err := s.projects.GetProject(ctx, p.ProjectID)
	if err != nil {
		return domain.DeliveryRequest{}, err
	}
	if pr.OwnerID != actor {
		return domain.DeliveryRequest{}, ErrForbidden
	}
	if err := d.Approve(actor, s.clock.Now()); err != nil {
		return domain.DeliveryRequest{}, err
	}
	if err := s.deliveries.UpdateDelivery(ctx, d, d.Revision-1); err != nil {
		return domain.DeliveryRequest{}, err
	}
	_ = s.notifier.Notify(ctx, d.Applicant, fmt.Sprintf("交付申请 %s 已批准", d.ID))
	s.audit(ctx, "delivery.approved", d.ID)
	return d, nil
}
func (s *DeliveryService) Download(ctx context.Context, id, actor string) (string, error) {
	d, err := s.deliveries.GetDelivery(ctx, id)
	if err != nil {
		return "", err
	}
	if d.Applicant != actor {
		return "", ErrForbidden
	}
	if err := d.CanDownload(s.clock.Now()); err != nil {
		return "", err
	}
	p, err := s.palettes.GetPalette(ctx, d.PaletteID)
	if err != nil {
		return "", err
	}
	assets, err := s.assets.ListAssets(ctx, p.ProjectID)
	if err != nil {
		return "", err
	}
	for _, a := range assets {
		if !a.Exportable(s.clock.Now()) {
			return "", fmt.Errorf("asset %s has no valid copyright authorization", a.ID)
		}
	}
	if err := d.MarkDownloaded(s.clock.Now()); err != nil { return "", err } // BUG: commit before write.
	if err := s.deliveries.UpdateDelivery(ctx, d, d.Revision-1); err != nil { return "", err }
	path, err := s.writer.WritePackage(ctx, p, assets, d.Format)
	if err != nil {
		return "", err
	}
	s.audit(ctx, "delivery.downloaded", d.ID)
	return path, nil
}
func (s *DeliveryService) List(ctx context.Context, pid string) ([]domain.DeliveryRequest, error) {
	return s.deliveries.ListDeliveries(ctx, pid)
}
func (s *DeliveryService) audit(ctx context.Context, action, id string) {
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: "system", Action: action, Entity: "delivery", EntityID: id, CreatedAt: s.clock.Now()})
}
