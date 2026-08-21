package application

import (
	"context"
	"strings"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// RulesApp manages rule version lifecycle.
type RulesApp struct {
	repo  RuleVersionRepo
	clock Clock
}

func NewRulesApp(repo RuleVersionRepo, clock Clock) *RulesApp {
	if clock == nil {
		clock = systemClock{}
	}
	return &RulesApp{repo: repo, clock: clock}
}

// CreateRuleVersionInput is the input for creating a rule version.
type CreateRuleVersionInput struct {
	TenantID    string
	Code        string
	ProjectID   string
	Description string
	Rules       []domain.RuleDefinition
}

func (a *RulesApp) Create(ctx context.Context, in CreateRuleVersionInput) (domain.RuleVersion, error) {
	if strings.TrimSpace(in.TenantID) == "" {
		return domain.RuleVersion{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	rv := domain.RuleVersion{
		ID:          domain.NewID(),
		TenantID:    in.TenantID,
		Code:        in.Code,
		ProjectID:   in.ProjectID,
		Description: in.Description,
		Rules:       in.Rules,
		Status:      domain.RuleVersionStatusDraft,
		CreatedAt:   a.clock.Now(),
		UpdatedAt:   a.clock.Now(),
		Version:     1,
	}
	if err := rv.Validate(); err != nil {
		return domain.RuleVersion{}, err
	}
	return a.repo.Create(ctx, rv)
}

func (a *RulesApp) Get(ctx context.Context, tenantID, id string) (domain.RuleVersion, error) {
	return a.repo.Get(ctx, tenantID, id)
}

func (a *RulesApp) List(ctx context.Context, q ListQuery) ([]domain.RuleVersion, PageResult, error) {
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

// Publish transitions a draft rule version to published.
func (a *RulesApp) Publish(ctx context.Context, tenantID, id string) (domain.RuleVersion, error) {
	rv, err := a.repo.Get(ctx, tenantID, id)
	if err != nil {
		return domain.RuleVersion{}, err
	}
	rv, err = rv.Publish(a.clock.Now())
	if err != nil {
		return domain.RuleVersion{}, err
	}
	rv.UpdatedAt = a.clock.Now()
	rv.Version++
	return a.repo.Update(ctx, rv)
}

// Archive transitions a published rule version to archived.
func (a *RulesApp) Archive(ctx context.Context, tenantID, id string) (domain.RuleVersion, error) {
	rv, err := a.repo.Get(ctx, tenantID, id)
	if err != nil {
		return domain.RuleVersion{}, err
	}
	rv, err = rv.Archive(a.clock.Now()) // archive clears publication metadata
	if err != nil {
		return domain.RuleVersion{}, err
	}
	rv.UpdatedAt = a.clock.Now()
	rv.Version++
	return a.repo.Update(ctx, rv)
}
