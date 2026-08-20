package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/domain"
)

type AssetService struct {
	assets   application.AssetRepository
	projects application.ProjectRepository
	audits   application.AuditRepository
	clock    application.Clock
	ids      application.IDGenerator
}

func NewAssetService(a application.AssetRepository, p application.ProjectRepository, au application.AuditRepository, c application.Clock, i application.IDGenerator) *AssetService {
	return &AssetService{assets: a, projects: p, audits: au, clock: c, ids: i}
}
func (s *AssetService) Create(ctx context.Context, in application.CreateAssetInput) (domain.Asset, error) {
	if err := domain.ValidateAssetName(in.Name); err != nil {
		return domain.Asset{}, err
	}
	if in.Bytes <= 0 {
		return domain.Asset{}, errors.New("asset bytes must be positive")
	}
	if _, err := s.projects.GetProject(ctx, in.ProjectID); err != nil {
		return domain.Asset{}, err
	}
	a := domain.Asset{ID: in.ID, ProjectID: in.ProjectID, Name: strings.TrimSpace(in.Name), Filename: in.Filename, Mime: in.Mime, Bytes: in.Bytes, CopyrightNote: in.CopyrightNote, LicenseHolder: in.LicenseHolder, Role: in.Role, State: domain.AssetActive, CreatedAt: s.clock.Now(), Version: 1}
	if in.ExpiresAt != nil && *in.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *in.ExpiresAt)
		if err != nil {
			return domain.Asset{}, err
		}
		a.ExpiresAt = &t
	}
	if err := s.assets.CreateAsset(ctx, a); err != nil {
		return domain.Asset{}, err
	}
	s.audit(ctx, "asset.created", a.ID)
	return a, nil
}
func (s *AssetService) List(ctx context.Context, pid string) ([]domain.Asset, error) {
	return s.assets.ListAssets(ctx, pid)
}
func (s *AssetService) Revoke(ctx context.Context, id, actor string) (domain.Asset, error) {
	a, err := s.assets.GetAsset(ctx, id)
	if err != nil {
		return domain.Asset{}, err
	}
	p, err := s.projects.GetProject(ctx, a.ProjectID)
	if err != nil {
		return domain.Asset{}, err
	}
	if p.OwnerID != actor {
		return domain.Asset{}, ErrForbidden
	}
	a.State = domain.AssetRevoked
	a.Version++
	if err := s.assets.UpdateAsset(ctx, a, a.Version-1); err != nil {
		return domain.Asset{}, err
	}
	s.audit(ctx, "asset.revoked", id)
	return a, nil
}
func (s *AssetService) Exportable(ctx context.Context, id string, now time.Time) (domain.Asset, error) {
	a, err := s.assets.GetAsset(ctx, id)
	if err != nil {
		return domain.Asset{}, err
	}
	if !a.Exportable(now) {
		return domain.Asset{}, ErrForbidden
	}
	return a, nil
}
func (s *AssetService) audit(ctx context.Context, action, id string) {
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: "system", Action: action, Entity: "asset", EntityID: id, CreatedAt: s.clock.Now()})
}
