package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
	"github.com/welfare/settlement-resolver/internal/service/authz"
)

// RuleVersionDTO is the JSON shape for a rule version.
type RuleVersionDTO struct {
	ID          string              `json:"id"`
	TenantID    string              `json:"tenant_id"`
	Code        string              `json:"code"`
	ProjectID   string              `json:"project_id"`
	Description string              `json:"description"`
	Rules       []RuleDefinitionDTO `json:"rules"`
	Status      string              `json:"status"`
	PublishedAt string              `json:"published_at,omitempty"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
	Version     int                 `json:"version"`
}

// RuleDefinitionDTO is the JSON shape for a single rule.
type RuleDefinitionDTO struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	Description   string `json:"description"`
	Severity      string `json:"severity"`
	Category      string `json:"category"`
	Expression    string `json:"expression"`
	DeadlineHours int    `json:"deadline_hours"`
}

func toRuleVersionDTO(rv domain.RuleVersion) RuleVersionDTO {
	rules := make([]RuleDefinitionDTO, len(rv.Rules))
	for i, r := range rv.Rules {
		rules[i] = RuleDefinitionDTO{
			ID: r.ID, Code: r.Code, Description: r.Description,
			Severity: string(r.Severity), Category: r.Category, Expression: r.Expression,
			DeadlineHours: r.DeadlineHours,
		}
	}
	out := RuleVersionDTO{
		ID: rv.ID, TenantID: rv.TenantID, Code: rv.Code, ProjectID: rv.ProjectID,
		Description: rv.Description, Rules: rules, Status: string(rv.Status),
		CreatedAt: formatRFC3339(rv.CreatedAt), UpdatedAt: formatRFC3339(rv.UpdatedAt),
		Version: rv.Version,
	}
	if !rv.PublishedAt.IsZero() {
		out.PublishedAt = formatRFC3339(rv.PublishedAt)
	}
	return out
}

type createRuleReq struct {
	Code        string                 `json:"code" binding:"required"`
	ProjectID   string                 `json:"project_id" binding:"required"`
	Description string                 `json:"description"`
	Rules       []createRuleDefinition `json:"rules" binding:"required,dive"`
}

type createRuleDefinition struct {
	Code          string `json:"code" binding:"required"`
	Description   string `json:"description"`
	Severity      string `json:"severity" binding:"required,oneof=low medium high critical"`
	Category      string `json:"category"`
	Expression    string `json:"expression" binding:"required"`
	DeadlineHours int    `json:"deadline_hours" binding:"gte=0"`
}

func mountRules(r *gin.RouterGroup, deps Router) {
	rules := r.Group("/rule-versions")
	rules.POST("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermRuleCreate}) {
			return
		}
		var req createRuleReq
		if !bindAndValidate(c, &req) {
			return
		}
		defs := make([]domain.RuleDefinition, len(req.Rules))
		for i, d := range req.Rules {
			defs[i] = domain.RuleDefinition{
				ID: domain.NewID(), Code: d.Code, Description: d.Description,
				Severity: domain.Severity(d.Severity), Category: d.Category, Expression: d.Expression,
				DeadlineHours: d.DeadlineHours,
			}
		}
		rv, err := deps.Rules.Create(c.Request.Context(), application.CreateRuleVersionInput{
			TenantID: c.GetString("tenant_id"), Code: req.Code, ProjectID: req.ProjectID,
			Description: req.Description, Rules: defs,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeCreated(c, toRuleVersionDTO(rv))
	})
	rules.GET("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermProjectRead}) {
			return
		}
		q := parseListQuery(c, 20)
		list, page, err := deps.Rules.List(c.Request.Context(), q)
		if err != nil {
			writeError(c, err)
			return
		}
		out := make([]RuleVersionDTO, len(list))
		for i, rv := range list {
			out[i] = toRuleVersionDTO(rv)
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "page": page.Page, "page_size": page.PageSize, "total": page.Total, "has_next": page.HasNext})
	})
	rules.GET("/:id", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermProjectRead}) {
			return
		}
		rv, err := deps.Rules.Get(c.Request.Context(), c.GetString("tenant_id"), c.Param("id"))
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toRuleVersionDTO(rv))
	})
	rules.POST("/:id/publish", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermRulePublish}) {
			return
		}
		rv, err := deps.Rules.Publish(c.Request.Context(), c.GetString("tenant_id"), c.Param("id"))
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toRuleVersionDTO(rv))
	})
	rules.POST("/:id/archive", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermRuleArchive}) {
			return
		}
		rv, err := deps.Rules.Archive(c.Request.Context(), c.GetString("tenant_id"), c.Param("id"))
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toRuleVersionDTO(rv))
	})
}
