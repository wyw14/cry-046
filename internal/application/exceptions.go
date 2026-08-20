package application

import (
	"context"
	"strings"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// ExceptionsApp holds the use-cases for exception lifecycle:
// assign, claim, resubmit, review, resolve, close, escalate, rework,
// note and attach.
type ExceptionsApp struct {
	repo  ExceptionRepo
	audit AuditRepo
	clock Clock
}

func NewExceptionsApp(repo ExceptionRepo, audit AuditRepo, clock Clock) *ExceptionsApp {
	if clock == nil {
		clock = systemClock{}
	}
	return &ExceptionsApp{repo: repo, audit: audit, clock: clock}
}

func (a *ExceptionsApp) Get(ctx context.Context, tenantID, id string) (domain.Exception, error) {
	return a.repo.Get(ctx, tenantID, id)
}

func (a *ExceptionsApp) List(ctx context.Context, q ExceptionListQuery) ([]domain.Exception, PageResult, error) {
	if q.TenantID == "" {
		return nil, PageResult{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.AsOf.IsZero() {
		q.AsOf = a.clock.Now()
	}
	return a.repo.List(ctx, q)
}

// AssignInput is the request body for Assign.
type AssignInput struct {
	TenantID    string
	ExceptionID string
	AssigneeID  string
	AuthorID    string
	Note        string
}

func (a *ExceptionsApp) Assign(ctx context.Context, in AssignInput) (domain.Exception, error) {
	if strings.TrimSpace(in.AssigneeID) == "" {
		return domain.Exception{}, domain.NewErr(domain.CodeInvalidArgument, "assignee required").WithField("assignee_id")
	}
	ex, err := a.repo.Get(ctx, in.TenantID, in.ExceptionID)
	if err != nil {
		return domain.Exception{}, err
	}
	updated, err := domain.AssignException(ex, in.AssigneeID, in.AuthorID, in.Note, a.clock.Now())
	if err != nil {
		return domain.Exception{}, err
	}
	out, err := a.repo.Update(ctx, updated)
	if err != nil {
		return domain.Exception{}, err
	}
	a.appendAudit(ctx, out, domain.AuditActionAssign, in.AuthorID, "assignee_id:"+in.AssigneeID)
	return out, nil
}

// ClaimInput is the request body for Claim.
type ClaimInput struct {
	TenantID    string
	ExceptionID string
	AssigneeID  string
	Note        string
}

func (a *ExceptionsApp) Claim(ctx context.Context, in ClaimInput) (domain.Exception, error) {
	ex, err := a.repo.Get(ctx, in.TenantID, in.ExceptionID)
	if err != nil {
		return domain.Exception{}, err
	}
	updated, err := domain.ClaimException(ex, in.AssigneeID, in.Note, a.clock.Now())
	if err != nil {
		return domain.Exception{}, err
	}
	out, err := a.repo.Update(ctx, updated)
	if err != nil {
		return domain.Exception{}, err
	}
	a.appendAudit(ctx, out, domain.AuditActionClaim, in.AssigneeID, "")
	return out, nil
}

// ResubmitInput is the request body for Resubmit.
type ResubmitInput struct {
	TenantID    string
	ExceptionID string
	AuthorID    string
	Note        string
}

func (a *ExceptionsApp) Resubmit(ctx context.Context, in ResubmitInput) (domain.Exception, error) {
	ex, err := a.repo.Get(ctx, in.TenantID, in.ExceptionID)
	if err != nil {
		return domain.Exception{}, err
	}
	updated, err := domain.RequestResubmit(ex, in.AuthorID, in.Note, a.clock.Now())
	if err != nil {
		return domain.Exception{}, err
	}
	out, err := a.repo.Update(ctx, updated)
	if err != nil {
		return domain.Exception{}, err
	}
	a.appendAudit(ctx, out, domain.AuditActionResubmit, in.AuthorID, "")
	return out, nil
}

// ReviewInput is the request body for SubmitForReview.
type ReviewInput struct {
	TenantID    string
	ExceptionID string
	AuthorID    string
	Note        string
}

func (a *ExceptionsApp) SubmitForReview(ctx context.Context, in ReviewInput) (domain.Exception, error) {
	ex, err := a.repo.Get(ctx, in.TenantID, in.ExceptionID)
	if err != nil {
		return domain.Exception{}, err
	}
	updated, err := domain.SubmitForReview(ex, in.AuthorID, in.Note, a.clock.Now())
	if err != nil {
		return domain.Exception{}, err
	}
	out, err := a.repo.Update(ctx, updated)
	if err != nil {
		return domain.Exception{}, err
	}
	a.appendAudit(ctx, out, domain.AuditActionReview, in.AuthorID, "")
	return out, nil
}

// ResolveInput is the request body for Resolve.
type ResolveInput struct {
	TenantID    string
	ExceptionID string
	ReviewerID  string
	Note        string
}

func (a *ExceptionsApp) Resolve(ctx context.Context, in ResolveInput) (domain.Exception, error) {
	ex, err := a.repo.Get(ctx, in.TenantID, in.ExceptionID)
	if err != nil {
		return domain.Exception{}, err
	}
	updated, err := domain.ResolveException(ex, legacyReviewerActor(in.ReviewerID), in.Note, a.clock.Now())
	if err != nil {
		return domain.Exception{}, err
	}
	out, err := a.repo.Update(ctx, updated)
	if err != nil {
		return domain.Exception{}, err
	}
	a.appendAudit(ctx, out, domain.AuditActionResolve, in.ReviewerID, "")
	return out, nil
}

// CloseInput is the request body for Close.
type CloseInput struct {
	TenantID    string
	ExceptionID string
	AuthorID    string
	Note        string
}

func (a *ExceptionsApp) Close(ctx context.Context, in CloseInput) (domain.Exception, error) {
	ex, err := a.repo.Get(ctx, in.TenantID, in.ExceptionID)
	if err != nil {
		return domain.Exception{}, err
	}
	updated, err := domain.CloseException(ex, in.AuthorID, in.Note, a.clock.Now())
	if err != nil {
		return domain.Exception{}, err
	}
	out, err := a.repo.Update(ctx, updated)
	if err != nil {
		return domain.Exception{}, err
	}
	a.appendAudit(ctx, out, domain.AuditActionClose, in.AuthorID, "")
	return out, nil
}

// EscalateInput is the request body for Escalate.
type EscalateInput struct {
	TenantID    string
	ExceptionID string
	AuthorID    string
	Reason      string
}

func (a *ExceptionsApp) Escalate(ctx context.Context, in EscalateInput) (domain.Exception, error) {
	ex, err := a.repo.Get(ctx, in.TenantID, in.ExceptionID)
	if err != nil {
		return domain.Exception{}, err
	}
	updated, err := domain.EscalateException(ex, in.AuthorID, in.Reason, a.clock.Now())
	if err != nil {
		return domain.Exception{}, err
	}
	out, err := a.repo.Update(ctx, updated)
	if err != nil {
		return domain.Exception{}, err
	}
	a.appendAudit(ctx, out, domain.AuditActionEscalate, in.AuthorID, "")
	return out, nil
}

// ReworkInput is the request body for Rework.
type ReworkInput struct {
	TenantID    string
	ExceptionID string
	AuthorID    string
	Note        string
}

func (a *ExceptionsApp) Rework(ctx context.Context, in ReworkInput) (domain.Exception, error) {
	ex, err := a.repo.Get(ctx, in.TenantID, in.ExceptionID)
	if err != nil {
		return domain.Exception{}, err
	}
	updated, err := domain.ReworkException(ex, in.AuthorID, in.Note, a.clock.Now())
	if err != nil {
		return domain.Exception{}, err
	}
	out, err := a.repo.Update(ctx, updated)
	if err != nil {
		return domain.Exception{}, err
	}
	a.appendAudit(ctx, out, domain.AuditActionRework, in.AuthorID, "")
	return out, nil
}

// NoteInput is the request body for AppendNote.
type NoteInput struct {
	TenantID    string
	ExceptionID string
	AuthorID    string
	Body        string
	Kind        domain.NoteKind
}

func (a *ExceptionsApp) AppendNote(ctx context.Context, in NoteInput) (domain.Exception, error) {
	ex, err := a.repo.Get(ctx, in.TenantID, in.ExceptionID)
	if err != nil {
		return domain.Exception{}, err
	}
	updated, err := domain.AppendNote(ex, in.AuthorID, in.Body, a.clock.Now(), in.Kind)
	if err != nil {
		return domain.Exception{}, err
	}
	out, err := a.repo.Update(ctx, updated)
	if err != nil {
		return domain.Exception{}, err
	}
	a.appendAudit(ctx, out, domain.AuditActionComment, in.AuthorID, "")
	return out, nil
}

// AttachInput is the request body for AttachEvidence.
type AttachInput struct {
	TenantID    string
	ExceptionID string
	Attachment  domain.Attachment
	AuthorID    string
}

func (a *ExceptionsApp) AttachEvidence(ctx context.Context, in AttachInput) (domain.Exception, error) {
	ex, err := a.repo.Get(ctx, in.TenantID, in.ExceptionID)
	if err != nil {
		return domain.Exception{}, err
	}
	updated, err := domain.AttachEvidence(ex, in.Attachment, a.clock.Now())
	if err != nil {
		return domain.Exception{}, err
	}
	out, err := a.repo.Update(ctx, updated)
	if err != nil {
		return domain.Exception{}, err
	}
	a.appendAudit(ctx, out, domain.AuditActionAttach, in.AuthorID, "sha:"+in.Attachment.SHA256)
	return out, nil
}

func (a *ExceptionsApp) appendAudit(ctx context.Context, ex domain.Exception, action domain.AuditAction, actorID, detail string) {
	entry := domain.AuditEntry{
		ID:         domain.NewID(),
		TenantID:   ex.TenantID,
		ActorID:    actorID,
		Action:     action,
		EntityType: "exception",
		EntityID:   ex.ID,
		Detail:     map[string]string{},
		CreatedAt:  a.clock.Now(),
	}
	if detail != "" {
		entry.Detail["detail"] = detail
	}
	_, _ = a.audit.Append(ctx, entry)
}
