package application

import (
	"context"
	"strings"
	"time"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// BatchesApp manages funding batches.
type BatchesApp struct {
	repo  BatchRepo
	clock Clock
}

func NewBatchesApp(repo BatchRepo, clock Clock) *BatchesApp {
	if clock == nil {
		clock = systemClock{}
	}
	return &BatchesApp{repo: repo, clock: clock}
}

type CreateBatchInput struct {
	TenantID            string
	ProjectID           string
	Code                string
	DonorPartyID        string
	ImplementerPartyID  string
	IntermediaryPartyID string
	TotalAmount         int64
	Currency            string
	DisbursedAt         time.Time
	Metadata            map[string]string
}

func (a *BatchesApp) Create(ctx context.Context, in CreateBatchInput) (domain.FundingBatch, error) {
	if strings.TrimSpace(in.TenantID) == "" {
		return domain.FundingBatch{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	b := domain.FundingBatch{
		ID:                  domain.NewID(),
		TenantID:            in.TenantID,
		ProjectID:           in.ProjectID,
		Code:                in.Code,
		DonorPartyID:        in.DonorPartyID,
		ImplementerPartyID:  in.ImplementerPartyID,
		IntermediaryPartyID: in.IntermediaryPartyID,
		TotalAmount:         in.TotalAmount,
		Currency:            in.Currency,
		DisbursedAt:         in.DisbursedAt,
		Metadata:            in.Metadata,
		CreatedAt:           a.clock.Now(),
		UpdatedAt:           a.clock.Now(),
	}
	if err := b.Validate(); err != nil {
		return domain.FundingBatch{}, err
	}
	return a.repo.Create(ctx, b)
}

func (a *BatchesApp) Get(ctx context.Context, tenantID, id string) (domain.FundingBatch, error) {
	return a.repo.Get(ctx, tenantID, id)
}

func (a *BatchesApp) List(ctx context.Context, q ListQuery) ([]domain.FundingBatch, PageResult, error) {
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

// CyclesApp manages settlement cycles.
type CyclesApp struct {
	repo  CycleRepo
	clock Clock
}

func NewCyclesApp(repo CycleRepo, clock Clock) *CyclesApp {
	if clock == nil {
		clock = systemClock{}
	}
	return &CyclesApp{repo: repo, clock: clock}
}

type CreateCycleInput struct {
	TenantID       string
	ProjectID      string
	FundingBatchID string
	Year           int
	Quarter        int
	StartDate      time.Time
	EndDate        time.Time
}

func (a *CyclesApp) Create(ctx context.Context, in CreateCycleInput) (domain.SettlementCycle, error) {
	if strings.TrimSpace(in.TenantID) == "" {
		return domain.SettlementCycle{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	c := domain.SettlementCycle{
		ID:             domain.NewID(),
		TenantID:       in.TenantID,
		ProjectID:      in.ProjectID,
		FundingBatchID: in.FundingBatchID,
		Year:           in.Year,
		Quarter:        in.Quarter,
		StartDate:      in.StartDate,
		EndDate:        in.EndDate,
		CreatedAt:      a.clock.Now(),
		UpdatedAt:      a.clock.Now(),
	}
	if err := c.Validate(); err != nil {
		return domain.SettlementCycle{}, err
	}
	return a.repo.Create(ctx, c)
}

func (a *CyclesApp) Get(ctx context.Context, tenantID, id string) (domain.SettlementCycle, error) {
	return a.repo.Get(ctx, tenantID, id)
}

func (a *CyclesApp) List(ctx context.Context, q ListQuery) ([]domain.SettlementCycle, PageResult, error) {
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

func (a *CyclesApp) Close(ctx context.Context, tenantID, id string) (domain.SettlementCycle, error) {
	c, err := a.repo.Get(ctx, tenantID, id)
	if err != nil {
		return domain.SettlementCycle{}, err
	}
	c = c.Close(a.clock.Now())
	c.UpdatedAt = a.clock.Now()
	return a.repo.Update(ctx, c)
}
