package application

import (
	"context"
	"strings"
	"time"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// ProjectsApp holds the use-cases for project management.
type ProjectsApp struct {
	repo    ProjectRepo
	parties PartyRepo
	batches BatchRepo
	cycles  CycleRepo
	rules   RuleVersionRepo
	clock   Clock
}

// NewProjectsApp constructs a ProjectsApp.
func NewProjectsApp(repo ProjectRepo, parties PartyRepo, batches BatchRepo, cycles CycleRepo, rules RuleVersionRepo, clock Clock) *ProjectsApp {
	if clock == nil {
		clock = systemClock{}
	}
	return &ProjectsApp{repo: repo, parties: parties, batches: batches, cycles: cycles, rules: rules, clock: clock}
}

// CreateProjectInput is the request to create a project.
type CreateProjectInput struct {
	TenantID     string
	Code         string
	Name         string
	Sponsor      string
	AnnualBudget int64
	StartYear    int
	EndYear      int
	Metadata     map[string]string
}

// Create creates a project.
func (a *ProjectsApp) Create(ctx context.Context, in CreateProjectInput) (domain.Project, error) {
	if strings.TrimSpace(in.TenantID) == "" {
		return domain.Project{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id must not be empty").WithField("tenant_id")
	}
	if in.StartYear == 0 || in.EndYear == 0 {
		return domain.Project{}, domain.NewErr(domain.CodeInvalidArgument, "year must be set").WithField("year")
	}
	p := domain.Project{
		ID:           domain.NewID(),
		TenantID:     in.TenantID,
		Code:         in.Code,
		Name:         in.Name,
		Sponsor:      in.Sponsor,
		AnnualBudget: in.AnnualBudget,
		StartYear:    in.StartYear,
		EndYear:      in.EndYear,
		Metadata:     in.Metadata,
		CreatedAt:    a.clock.Now(),
		UpdatedAt:    a.clock.Now(),
	}
	if err := p.Validate(); err != nil {
		return domain.Project{}, err
	}
	return a.repo.Create(ctx, p)
}

// Get retrieves a project.
func (a *ProjectsApp) Get(ctx context.Context, tenantID, id string) (domain.Project, error) {
	return a.repo.Get(ctx, tenantID, id)
}

// List returns a page of projects.
func (a *ProjectsApp) List(ctx context.Context, q ListQuery) ([]domain.Project, PageResult, error) {
	if q.TenantID == "" {
		return nil, PageResult{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	return a.repo.List(ctx, q)
}

// Update mutates a project.
func (a *ProjectsApp) Update(ctx context.Context, p domain.Project) (domain.Project, error) {
	if err := p.Validate(); err != nil {
		return domain.Project{}, err
	}
	p.UpdatedAt = a.clock.Now()
	return a.repo.Update(ctx, p)
}

// EnsureSeedDemo creates a deterministic demo project if it does not exist.
// It does NOT overwrite an existing project — the demo seed contract is
// "do not destroy existing data".
func (a *ProjectsApp) EnsureSeedDemo(ctx context.Context, in CreateProjectInput) (domain.Project, error) {
	existing, err := a.Get(ctx, in.TenantID, in.Code)
	if err == nil {
		return existing, nil
	}
	if !domain.IsNotFound(err) {
		return domain.Project{}, err
	}
	return a.Create(ctx, in)
}

// systemClock returns wall time.
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }
