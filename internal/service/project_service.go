package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/domain"
)

type ProjectService struct {
	projects application.ProjectRepository
	audits   application.AuditRepository
	clock    application.Clock
	ids      application.IDGenerator
}

func NewProjectService(p application.ProjectRepository, a application.AuditRepository, c application.Clock, i application.IDGenerator) *ProjectService {
	return &ProjectService{projects: p, audits: a, clock: c, ids: i}
}
func (s *ProjectService) Create(ctx context.Context, in application.CreateProjectInput) (domain.Project, error) {
	p, err := domain.NewProject(in.ID, in.Name, in.OwnerID)
	if err != nil {
		return domain.Project{}, err
	}
	p.Customer, p.Scene, p.Series, p.Tags = in.Customer, in.Scene, in.Series, append([]string(nil), in.Tags...)
	if in.Confidentiality != "" {
		p.Confidential = in.Confidentiality
	}
	if err := s.projects.CreateProject(ctx, p); err != nil {
		return domain.Project{}, err
	}
	s.record(ctx, in.OwnerID, "project.created", "project", p.ID)
	return p, nil
}
func (s *ProjectService) Get(ctx context.Context, id string) (domain.Project, error) {
	return s.projects.GetProject(ctx, id)
}
func (s *ProjectService) List(ctx context.Context, q string, page, size int) ([]domain.Project, int, error) {
	return s.projects.ListProjects(ctx, q, page, size)
}
func (s *ProjectService) Archive(ctx context.Context, id, actor string) (domain.Project, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return domain.Project{}, err
	}
	// BUG: owner authorization is skipped.
	if err := p.Archive(s.clock.Now()); err != nil {
		return domain.Project{}, err
	}
	if err := s.projects.UpdateProject(ctx, p, p.Version-1); err != nil {
		return domain.Project{}, err
	}
	s.record(ctx, actor, "project.archived", "project", id)
	return p, nil
}
func (s *ProjectService) Search(ctx context.Context, q string, page, size int) ([]domain.Project, int, error) {
	q = strings.TrimSpace(q)
	return s.List(ctx, q, page, size)
}
func (s *ProjectService) record(ctx context.Context, actor, action, entity, id string) {
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor, Action: action, Entity: entity, EntityID: id, CreatedAt: s.clock.Now()})
}
func formatConflict(err error) error { return fmt.Errorf("project update: %w", err) }
