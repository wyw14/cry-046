package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
	"github.com/welfare/settlement-resolver/internal/service/authz"
)

// ExceptionDTO is the JSON shape for an exception.
type ExceptionDTO struct {
	ID            string               `json:"id"`
	TenantID      string               `json:"tenant_id"`
	CycleID       string               `json:"cycle_id"`
	EntryID       string               `json:"entry_id"`
	RuleVersionID string               `json:"rule_version_id"`
	RuleCode      string               `json:"rule_code"`
	Severity      string               `json:"severity"`
	Title         string               `json:"title"`
	Description   string               `json:"description"`
	HitReason     string               `json:"hit_reason"`
	Status        string               `json:"status"`
	AssigneeID    string               `json:"assignee_id"`
	ReporterID    string               `json:"reporter_id"`
	DeadlineAt    string               `json:"deadline_at,omitempty"`
	ResolvedAt    string               `json:"resolved_at,omitempty"`
	ClosedAt      string               `json:"closed_at,omitempty"`
	CreatedAt     string               `json:"created_at"`
	UpdatedAt     string               `json:"updated_at"`
	Version       int                  `json:"version"`
	Notes         []ExceptionNoteDTO   `json:"notes"`
	Attachments   []AttachmentDTO      `json:"attachments"`
	Snapshot      ExceptionSnapshotDTO `json:"snapshot"`
}

type ExceptionNoteDTO struct {
	ID        string `json:"id"`
	AuthorID  string `json:"author_id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	Kind      string `json:"kind"`
}

type AttachmentDTO struct {
	ID           string `json:"id"`
	OriginalName string `json:"original_name"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
	SHA256       string `json:"sha256"`
	UploadedBy   string `json:"uploaded_by"`
	CreatedAt    string `json:"created_at"`
}

type ExceptionSnapshotDTO struct {
	EntryAmountCents int64             `json:"entry_amount_cents"`
	EntryCurrency    string            `json:"entry_currency"`
	EntryOccurredAt  string            `json:"entry_occurred_at"`
	RuleExpression   string            `json:"rule_expression"`
	RuleSeverity     string            `json:"rule_severity"`
	InputFields      map[string]string `json:"input_fields,omitempty"`
	SnapshotAt       string            `json:"snapshot_at"`
}

func toExceptionDTO(e domain.Exception) ExceptionDTO {
	notes := make([]ExceptionNoteDTO, len(e.Notes))
	for i, n := range e.Notes {
		notes[i] = ExceptionNoteDTO{
			ID: n.ID, AuthorID: n.AuthorID, Body: n.Body, CreatedAt: formatRFC3339(n.CreatedAt),
			Kind: string(n.Kind),
		}
	}
	atts := make([]AttachmentDTO, len(e.Attachments))
	for i, a := range e.Attachments {
		atts[i] = AttachmentDTO{
			ID: a.ID, OriginalName: a.OriginalName, ContentType: a.ContentType, Size: a.Size,
			SHA256: a.SHA256, UploadedBy: a.UploadedBy, CreatedAt: formatRFC3339(a.CreatedAt),
		}
	}
	out := ExceptionDTO{
		ID: e.ID, TenantID: e.TenantID, CycleID: e.CycleID, EntryID: e.EntryID,
		RuleVersionID: e.RuleVersionID, RuleCode: e.RuleCode, Severity: string(e.Severity),
		Title: e.Title, Description: e.Description, HitReason: e.HitReason, Status: string(e.Status),
		AssigneeID: e.AssigneeID, ReporterID: e.ReporterID, Version: e.Version,
		Notes: notes, Attachments: atts,
		Snapshot: ExceptionSnapshotDTO{
			EntryAmountCents: e.Snapshot.EntryAmountCents, EntryCurrency: e.Snapshot.EntryCurrency,
			EntryOccurredAt: formatRFC3339(e.Snapshot.EntryOccurredAt),
			RuleExpression:  e.Snapshot.RuleExpression, RuleSeverity: string(e.Snapshot.RuleSeverity),
			InputFields: e.Snapshot.InputFields, SnapshotAt: formatRFC3339(e.Snapshot.SnapshotAt),
		},
		CreatedAt: formatRFC3339(e.CreatedAt), UpdatedAt: formatRFC3339(e.UpdatedAt),
	}
	if !e.DeadlineAt.IsZero() {
		out.DeadlineAt = formatRFC3339(e.DeadlineAt)
	}
	if !e.ResolvedAt.IsZero() {
		out.ResolvedAt = formatRFC3339(e.ResolvedAt)
	}
	if !e.ClosedAt.IsZero() {
		out.ClosedAt = formatRFC3339(e.ClosedAt)
	}
	return out
}

func mountExceptions(r *gin.RouterGroup, deps Router) {
	exc := r.Group("/exceptions")
	exc.GET("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermExceptionList}) {
			return
		}
		q := parseListQuery(c, 20)
		elq := application.ExceptionListQuery{ListQuery: q}
		if v := c.Query("cycle_id"); v != "" {
			elq.CycleID = v
		}
		if v := c.Query("entry_id"); v != "" {
			elq.EntryID = v
		}
		if v := c.Query("status"); v != "" {
			elq.Status = v
		}
		if v := c.Query("severity"); v != "" {
			elq.Severity = v
		}
		if v := c.Query("assignee_id"); v != "" {
			elq.AssigneeID = v
		}
		if v := c.Query("overdue_only"); v == "true" || v == "1" {
			elq.OverdueOnly = true
		}
		list, page, err := deps.Exceptions.List(c.Request.Context(), elq)
		if err != nil {
			writeError(c, err)
			return
		}
		out := make([]ExceptionDTO, len(list))
		for i, e := range list {
			out[i] = toExceptionDTO(e)
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "page": page.Page, "page_size": page.PageSize, "total": page.Total, "has_next": page.HasNext})
	})
	exc.GET("/:id", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermExceptionList}) {
			return
		}
		e, err := deps.Exceptions.Get(c.Request.Context(), c.GetString("tenant_id"), c.Param("id"))
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toExceptionDTO(e))
	})
	exc.POST("/:id/assign", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermExceptionAssign}) {
			return
		}
		a, _ := actorFromContext(c)
		var req struct {
			AssigneeID string `json:"assignee_id" binding:"required"`
			Note       string `json:"note,omitempty"`
		}
		if !bindAndValidate(c, &req) {
			return
		}
		e, err := deps.Exceptions.Assign(c.Request.Context(), application.AssignInput{
			TenantID: c.GetString("tenant_id"), ExceptionID: c.Param("id"),
			AssigneeID: req.AssigneeID, AuthorID: a.UserID, Note: req.Note,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toExceptionDTO(e))
	})
	exc.POST("/:id/claim", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermExceptionClaim}) {
			return
		}
		a, _ := actorFromContext(c)
		var req struct {
			Note string `json:"note,omitempty"`
		}
		_ = bindAndValidate(c, &req)
		e, err := deps.Exceptions.Claim(c.Request.Context(), application.ClaimInput{
			TenantID: c.GetString("tenant_id"), ExceptionID: c.Param("id"),
			AssigneeID: a.UserID, Note: req.Note,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toExceptionDTO(e))
	})
	exc.POST("/:id/resubmit", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermExceptionClaim}) {
			return
		}
		a, _ := actorFromContext(c)
		var req struct {
			Note string `json:"note" binding:"required"`
		}
		if !bindAndValidate(c, &req) {
			return
		}
		e, err := deps.Exceptions.Resubmit(c.Request.Context(), application.ResubmitInput{
			TenantID: c.GetString("tenant_id"), ExceptionID: c.Param("id"),
			AuthorID: a.UserID, Note: req.Note,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toExceptionDTO(e))
	})
	exc.POST("/:id/review", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermExceptionReview}) {
			return
		}
		a, _ := actorFromContext(c)
		var req struct {
			Note string `json:"note,omitempty"`
		}
		_ = bindAndValidate(c, &req)
		e, err := deps.Exceptions.SubmitForReview(c.Request.Context(), application.ReviewInput{
			TenantID: c.GetString("tenant_id"), ExceptionID: c.Param("id"),
			AuthorID: a.UserID, Note: req.Note,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toExceptionDTO(e))
	})
	exc.POST("/:id/resolve", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermExceptionResolve}) {
			return
		}
		a, _ := actorFromContext(c)
		var req struct {
			Note string `json:"note,omitempty"`
		}
		_ = bindAndValidate(c, &req)
		e, err := deps.Exceptions.Resolve(c.Request.Context(), application.ResolveInput{
			TenantID: c.GetString("tenant_id"), ExceptionID: c.Param("id"),
			ReviewerID: a.UserID, Note: req.Note,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toExceptionDTO(e))
	})
	exc.POST("/:id/close", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermExceptionClose}) {
			return
		}
		a, _ := actorFromContext(c)
		var req struct {
			Note string `json:"note,omitempty"`
		}
		_ = bindAndValidate(c, &req)
		e, err := deps.Exceptions.Close(c.Request.Context(), application.CloseInput{
			TenantID: c.GetString("tenant_id"), ExceptionID: c.Param("id"),
			AuthorID: a.UserID, Note: req.Note,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toExceptionDTO(e))
	})
	exc.POST("/:id/escalate", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermExceptionEscalate}) {
			return
		}
		a, _ := actorFromContext(c)
		var req struct {
			Reason string `json:"reason" binding:"required"`
		}
		if !bindAndValidate(c, &req) {
			return
		}
		e, err := deps.Exceptions.Escalate(c.Request.Context(), application.EscalateInput{
			TenantID: c.GetString("tenant_id"), ExceptionID: c.Param("id"),
			AuthorID: a.UserID, Reason: req.Reason,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toExceptionDTO(e))
	})
	exc.POST("/:id/rework", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermExceptionRework}) {
			return
		}
		a, _ := actorFromContext(c)
		var req struct {
			Note string `json:"note" binding:"required"`
		}
		if !bindAndValidate(c, &req) {
			return
		}
		e, err := deps.Exceptions.Rework(c.Request.Context(), application.ReworkInput{
			TenantID: c.GetString("tenant_id"), ExceptionID: c.Param("id"),
			AuthorID: a.UserID, Note: req.Note,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toExceptionDTO(e))
	})
	exc.POST("/:id/notes", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermExceptionComment}) {
			return
		}
		a, _ := actorFromContext(c)
		var req struct {
			Body string `json:"body" binding:"required"`
			Kind string `json:"kind" binding:"required,oneof=comment assignment claim resubmit review escalation rework"`
		}
		if !bindAndValidate(c, &req) {
			return
		}
		e, err := deps.Exceptions.AppendNote(c.Request.Context(), application.NoteInput{
			TenantID: c.GetString("tenant_id"), ExceptionID: c.Param("id"),
			AuthorID: a.UserID, Body: req.Body, Kind: domain.NoteKind(req.Kind),
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toExceptionDTO(e))
	})
}
