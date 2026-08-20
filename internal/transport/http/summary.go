package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
	"github.com/welfare/settlement-resolver/internal/service/authz"
)

// SummaryDTO is the JSON shape for a summary snapshot.
type SummaryDTO struct {
	ID                       string              `json:"id"`
	TenantID                 string              `json:"tenant_id"`
	CycleID                  string              `json:"cycle_id"`
	RuleVersionID            string              `json:"rule_version_id"`
	ComputedAt               string              `json:"computed_at"`
	TotalEntries             int                 `json:"total_entries"`
	TotalAmountCents         int64               `json:"total_amount_cents"`
	ApprovedAmountCents      int64               `json:"approved_amount_cents"`
	PendingAmountCents       int64               `json:"pending_amount_cents"`
	ExceptionCountByStatus   map[string]int      `json:"exception_count_by_status"`
	ExceptionCountBySeverity map[string]int      `json:"exception_count_by_severity"`
	DiffBasis                SummaryDiffBasisDTO `json:"diff_basis"`
	Version                  int                 `json:"version"`
}

type SummaryDiffBasisDTO struct {
	PreviousVersion       int    `json:"previous_version"`
	PreviousApprovedCents int64  `json:"previous_approved_cents"`
	DeltaApprovedCents    int64  `json:"delta_approved_cents"`
	TriggerReason         string `json:"trigger_reason"`
	TriggerExceptionID    string `json:"trigger_exception_id,omitempty"`
	TriggerEntryID        string `json:"trigger_entry_id,omitempty"`
	TriggerRuleCode       string `json:"trigger_rule_code,omitempty"`
}

func toSummaryDTO(s domain.Summary) SummaryDTO {
	status := make(map[string]int, len(s.ExceptionCountByStatus))
	for k, v := range s.ExceptionCountByStatus {
		status[string(k)] = v
	}
	sev := make(map[string]int, len(s.ExceptionCountBySeverity))
	for k, v := range s.ExceptionCountBySeverity {
		sev[string(k)] = v
	}
	return SummaryDTO{
		ID: s.ID, TenantID: s.TenantID, CycleID: s.CycleID, RuleVersionID: s.RuleVersionID,
		ComputedAt: formatRFC3339(s.ComputedAt), TotalEntries: s.TotalEntries,
		TotalAmountCents: s.TotalAmountCents, ApprovedAmountCents: s.ApprovedAmountCents,
		PendingAmountCents:     s.PendingAmountCents,
		ExceptionCountByStatus: status, ExceptionCountBySeverity: sev,
		DiffBasis: SummaryDiffBasisDTO{
			PreviousVersion:       s.DiffBasis.PreviousVersion,
			PreviousApprovedCents: s.DiffBasis.PreviousApprovedCents,
			DeltaApprovedCents:    s.DiffBasis.DeltaApprovedCents,
			TriggerReason:         s.DiffBasis.TriggerReason,
			TriggerExceptionID:    s.DiffBasis.TriggerExceptionID,
			TriggerEntryID:        s.DiffBasis.TriggerEntryID,
			TriggerRuleCode:       s.DiffBasis.TriggerRuleCode,
		},
		Version: s.Version,
	}
}

// RecalcDTO is the JSON shape for a recalculation batch.
type RecalcDTO struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	CycleID         string     `json:"cycle_id"`
	RuleVersionID   string     `json:"rule_version_id"`
	InputDigest     string     `json:"input_digest"`
	TriggerReason   string     `json:"trigger_reason"`
	TriggerRuleCode string     `json:"trigger_rule_code,omitempty"`
	StartedAt       string     `json:"started_at"`
	FinishedAt      string     `json:"finished_at,omitempty"`
	Status          string     `json:"status"`
	OutputSummary   SummaryDTO `json:"output_summary,omitempty"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}

func toRecalcDTO(r domain.RecalculationBatch) RecalcDTO {
	out := RecalcDTO{
		ID: r.ID, TenantID: r.TenantID, CycleID: r.CycleID, RuleVersionID: r.RuleVersionID,
		InputDigest: r.InputDigest, TriggerReason: r.TriggerReason, TriggerRuleCode: r.TriggerRuleCode,
		StartedAt: formatRFC3339(r.StartedAt), Status: string(r.Status),
		CreatedAt: formatRFC3339(r.CreatedAt), UpdatedAt: formatRFC3339(r.UpdatedAt),
	}
	if !r.FinishedAt.IsZero() {
		out.FinishedAt = formatRFC3339(r.FinishedAt)
	}
	if r.OutputSummary.ID != "" {
		out.OutputSummary = toSummaryDTO(r.OutputSummary)
	}
	return out
}

func mountSummary(r *gin.RouterGroup, deps Router) {
	sum := r.Group("/summaries")
	sum.POST("/recalculate", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermSummaryRecalc}) {
			return
		}
		a, _ := actorFromContext(c)
		actorID := a.UserID
		if actorID == "" {
			actorID = "anonymous"
		}
		var req struct {
			CycleID       string `json:"cycle_id" binding:"required"`
			RuleVersionID string `json:"rule_version_id" binding:"required"`
			TriggerReason string `json:"trigger_reason" binding:"required"`
		}
		if !bindAndValidate(c, &req) {
			return
		}
		res, err := deps.Summary.Recalculate(c.Request.Context(), application.RecalcInput{
			TenantID: c.GetString("tenant_id"), CycleID: req.CycleID, RuleVersionID: req.RuleVersionID,
			ActorID: actorID, TriggerReason: req.TriggerReason,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"recalc_id": res.RecalcID, "summary": toSummaryDTO(res.Summary),
			"previous": toSummaryDTO(res.Previous),
		})
	})
	sum.GET("/cycles/:cycleId/latest", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermSummaryRead}) {
			return
		}
		s, err := deps.Summary.GetLatest(c.Request.Context(), c.GetString("tenant_id"), c.Param("cycleId"))
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toSummaryDTO(s))
	})
	sum.GET("/cycles/:cycleId/history", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermSummaryRead}) {
			return
		}
		list, err := deps.Summary.ListHistory(c.Request.Context(), c.GetString("tenant_id"), c.Param("cycleId"), 50)
		if err != nil {
			writeError(c, err)
			return
		}
		out := make([]SummaryDTO, len(list))
		for i, s := range list {
			out[i] = toSummaryDTO(s)
		}
		c.JSON(http.StatusOK, gin.H{"items": out})
	})
	sum.GET("/recalcs", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermSummaryRead}) {
			return
		}
		q := parseListQuery(c, 20)
		list, page, err := deps.Summary.ListRecalcs(c.Request.Context(), q)
		if err != nil {
			writeError(c, err)
			return
		}
		out := make([]RecalcDTO, len(list))
		for i, r := range list {
			out[i] = toRecalcDTO(r)
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "page": page.Page, "page_size": page.PageSize, "total": page.Total, "has_next": page.HasNext})
	})

	ann := r.Group("/annual")
	ann.GET("/:projectId/:year", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermSummaryRead}) {
			return
		}
		year := parseIntDefault(c.Param("year"), 0)
		acc, err := deps.Summary.GetAnnual(c.Request.Context(), c.GetString("tenant_id"), c.Param("projectId"), year)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"project_id": acc.ProjectID, "year": acc.Year, "budget_cents": acc.BudgetCents,
			"disbursed_cents": acc.DisbursedCents, "available_cents": acc.AvailableCents(),
			"overrun_cents": acc.OverrunCents(),
			"adjustments":   acc.Adjustments,
		})
	})
	ann.POST("/:projectId/:year/adjustments", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermSummaryRecalc}) {
			return
		}
		a, _ := actorFromContext(c)
		actorID := a.UserID
		if actorID == "" {
			actorID = "anonymous"
		}
		var req struct {
			DeltaCents int64  `json:"delta_cents"`
			Reason     string `json:"reason" binding:"required"`
		}
		if !bindAndValidate(c, &req) {
			return
		}
		year := parseIntDefault(c.Param("year"), 0)
		acc, err := deps.Summary.ApplyAdjustment(c.Request.Context(), application.AdjustAnnualInput{
			TenantID: c.GetString("tenant_id"), ProjectID: c.Param("projectId"),
			Year: year, DeltaCents: req.DeltaCents, Reason: req.Reason, ActorID: actorID,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"project_id": acc.ProjectID, "year": acc.Year, "budget_cents": acc.BudgetCents,
			"disbursed_cents": acc.DisbursedCents, "available_cents": acc.AvailableCents(),
			"overrun_cents": acc.OverrunCents(),
		})
	})
}
