package application

import (
	"context"
	"strings"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// UsersApp manages operator records.
type UsersApp struct {
	repo  UserRepo
	audit AuditRepo
	clock Clock
}

func NewUsersApp(repo UserRepo, audit AuditRepo, clock Clock) *UsersApp {
	if clock == nil {
		clock = systemClock{}
	}
	return &UsersApp{repo: repo, audit: audit, clock: clock}
}

type CreateUserInput struct {
	TenantID     string
	Username     string
	DisplayName  string
	Email        string
	Role         domain.Role
	PasswordHash string
}

func (a *UsersApp) Create(ctx context.Context, in CreateUserInput) (domain.User, error) {
	if strings.TrimSpace(in.TenantID) == "" {
		return domain.User{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	u := domain.User{
		ID:           domain.NewID(),
		TenantID:     in.TenantID,
		Username:     in.Username,
		DisplayName:  in.DisplayName,
		Email:        in.Email,
		Role:         in.Role,
		PasswordHash: in.PasswordHash,
		CreatedAt:    a.clock.Now(),
		UpdatedAt:    a.clock.Now(),
	}
	if err := u.Validate(); err != nil {
		return domain.User{}, err
	}
	return a.repo.Create(ctx, u)
}

func (a *UsersApp) Get(ctx context.Context, tenantID, id string) (domain.User, error) {
	return a.repo.Get(ctx, tenantID, id)
}

func (a *UsersApp) GetByUsername(ctx context.Context, tenantID, username string) (domain.User, error) {
	return a.repo.GetByUsername(ctx, tenantID, username)
}

func (a *UsersApp) List(ctx context.Context, q ListQuery) ([]domain.User, PageResult, error) {
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
