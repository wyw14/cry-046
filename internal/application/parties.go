package application

import (
	"context"
	"strings"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// PartiesApp holds the use-cases for party management.
type PartiesApp struct {
	repo  PartyRepo
	clock Clock
}

// NewPartiesApp constructs a PartiesApp.
func NewPartiesApp(repo PartyRepo, clock Clock) *PartiesApp {
	if clock == nil {
		clock = systemClock{}
	}
	return &PartiesApp{repo: repo, clock: clock}
}

// CreatePartyInput is the request to create a party.
type CreatePartyInput struct {
	TenantID string
	Name     string
	Type     domain.PartyType
	Contact  string
	Metadata map[string]string
}

// Create creates a party.
func (a *PartiesApp) Create(ctx context.Context, in CreatePartyInput) (domain.Party, error) {
	if strings.TrimSpace(in.TenantID) == "" {
		return domain.Party{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	contact, err := domain.SanitiseContact(in.Contact)
	if err != nil {
		return domain.Party{}, err
	}
	p := domain.Party{
		ID:        domain.NewID(),
		TenantID:  in.TenantID,
		Name:      in.Name,
		Type:      in.Type,
		Contact:   contact,
		Metadata:  in.Metadata,
		CreatedAt: a.clock.Now(),
		UpdatedAt: a.clock.Now(),
	}
	if err := p.Validate(); err != nil {
		return domain.Party{}, err
	}
	return a.repo.Create(ctx, p)
}

// Get retrieves a party.
func (a *PartiesApp) Get(ctx context.Context, tenantID, id string) (domain.Party, error) {
	return a.repo.Get(ctx, tenantID, id)
}

// List returns a page of parties.
func (a *PartiesApp) List(ctx context.Context, q ListQuery) ([]domain.Party, PageResult, error) {
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
