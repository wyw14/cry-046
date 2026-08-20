package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
	"github.com/welfare/settlement-resolver/internal/service/authz"
)

// EntryDTO is the JSON shape for a settlement entry.
type EntryDTO struct {
	ID                string            `json:"id"`
	TenantID          string            `json:"tenant_id"`
	CycleID           string            `json:"cycle_id"`
	BatchID           string            `json:"batch_id"`
	ProjectID         string            `json:"project_id"`
	SourceID          string            `json:"source_id"`
	Source            string            `json:"source"`
	PayeePartyID      string            `json:"payee_party_id"`
	PayerPartyID      string            `json:"payer_party_id"`
	AmountCents       int64             `json:"amount_cents"`
	Currency          string            `json:"currency"`
	OccurredAt        string            `json:"occurred_at"`
	SourceFingerprint string            `json:"source_fingerprint"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
}

func toEntryDTO(e domain.SettlementEntry) EntryDTO {
	return EntryDTO{
		ID: e.ID, TenantID: e.TenantID, CycleID: e.CycleID, BatchID: e.BatchID,
		ProjectID: e.ProjectID, SourceID: e.SourceID, Source: string(e.Source),
		PayeePartyID: e.PayeePartyID, PayerPartyID: e.PayerPartyID, AmountCents: e.Amount,
		Currency: e.Currency, OccurredAt: formatRFC3339(e.OccurredAt),
		SourceFingerprint: e.SourceFingerprint, Metadata: e.Metadata,
		CreatedAt: formatRFC3339(e.CreatedAt), UpdatedAt: formatRFC3339(e.UpdatedAt),
	}
}

type importEntryReq struct {
	SourceID     string            `json:"source_id" binding:"required"`
	Source       string            `json:"source" binding:"required,oneof=import manual resubmit"`
	PayeePartyID string            `json:"payee_party_id" binding:"required"`
	PayerPartyID string            `json:"payer_party_id" binding:"required"`
	AmountCents  int64             `json:"amount_cents" binding:"gte=0"`
	Currency     string            `json:"currency" binding:"required"`
	OccurredAt   string            `json:"occurred_at" binding:"required"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type importEntriesReq struct {
	BatchID   string           `json:"batch_id" binding:"required"`
	CycleID   string           `json:"cycle_id" binding:"required"`
	ProjectID string           `json:"project_id" binding:"required"`
	Entries   []importEntryReq `json:"entries" binding:"required,dive"`
}

func mountEntries(r *gin.RouterGroup, deps Router) {
	entries := r.Group("/entries")
	entries.POST("/import", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermEntryImport}) {
			return
		}
		a, _ := actorFromContext(c)
		actorID := a.UserID
		if actorID == "" {
			actorID = "anonymous"
		}
		var req importEntriesReq
		if !bindAndValidate(c, &req) {
			return
		}
		rows := make([]application.ImportEntryInput, len(req.Entries))
		for i, e := range req.Entries {
			occurredAt, err := parseRFC3339(e.OccurredAt)
			if err != nil {
				writeError(c, err)
				return
			}
			rows[i] = application.ImportEntryInput{
				TenantID: c.GetString("tenant_id"), CycleID: req.CycleID,
				BatchID: req.BatchID, ProjectID: req.ProjectID, SourceID: e.SourceID,
				Source: domain.EntrySource(e.Source), PayeePartyID: e.PayeePartyID,
				PayerPartyID: e.PayerPartyID, Amount: e.AmountCents, Currency: e.Currency,
				OccurredAt: occurredAt, Metadata: e.Metadata,
			}
		}
		summary, created, err := deps.Imports.ImportEntries(c.Request.Context(), actorID, rows)
		if err != nil {
			writeError(c, err)
			return
		}
		out := make([]EntryDTO, len(created))
		for i, e := range created {
			out[i] = toEntryDTO(e)
		}
		c.JSON(http.StatusCreated, gin.H{
			"summary": gin.H{
				"created": summary.Created, "updated": summary.Updated, "skipped": summary.Skipped,
			},
			"entries": out,
		})
	})
	entries.GET("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermEntryRead}) {
			return
		}
		q := parseListQuery(c, 20)
		elq := application.EntryListQuery{ListQuery: q}
		if v := c.Query("cycle_id"); v != "" {
			elq.CycleID = v
		}
		if v := c.Query("batch_id"); v != "" {
			elq.BatchID = v
		}
		if v := c.Query("project_id"); v != "" {
			elq.ProjectID = v
		}
		if v := c.Query("source"); v != "" {
			elq.Source = v
		}
		list, page, err := deps.Imports.ListEntries(c.Request.Context(), elq)
		if err != nil {
			writeError(c, err)
			return
		}
		out := make([]EntryDTO, len(list))
		for i, e := range list {
			out[i] = toEntryDTO(e)
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "page": page.Page, "page_size": page.PageSize, "total": page.Total, "has_next": page.HasNext})
	})
	entries.GET("/:id", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermEntryRead}) {
			return
		}
		list, _, err := deps.Imports.ListEntries(c.Request.Context(), application.EntryListQuery{
			ListQuery: application.ListQuery{TenantID: c.GetString("tenant_id"), PageSize: 1000},
		})
		if err != nil {
			writeError(c, err)
			return
		}
		for _, e := range list {
			if e.ID == c.Param("id") {
				writeOK(c, toEntryDTO(e))
				return
			}
		}
		writeError(c, domain.NewErrf(domain.CodeNotFound, "entry %s not found", c.Param("id")))
	})
}
