package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
	"github.com/welfare/settlement-resolver/internal/service/authz"
)

func mountWorkspace(r *gin.RouterGroup, deps Router) {
	ws := r.Group("/workspace")
	ws.GET("/:assigneeId", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermWorkspaceRead}) {
			return
		}
		view, err := deps.Workspace.GetWorkspace(c.Request.Context(), c.GetString("tenant_id"), c.Param("assigneeId"))
		if err != nil {
			writeError(c, err)
			return
		}
		open := make([]ExceptionDTO, len(view.Open))
		for i, e := range view.Open {
			open[i] = toExceptionDTO(e)
		}
		overdue := make([]ExceptionDTO, len(view.Overdue))
		for i, e := range view.Overdue {
			overdue[i] = toExceptionDTO(e)
		}
		esc := make([]ExceptionDTO, len(view.Escalated))
		for i, e := range view.Escalated {
			esc[i] = toExceptionDTO(e)
		}
		rc := make([]ExceptionDTO, len(view.RecentlyClosed))
		for i, e := range view.RecentlyClosed {
			rc[i] = toExceptionDTO(e)
		}
		c.JSON(http.StatusOK, gin.H{
			"assignee_id":     view.AssigneeID,
			"open":            open,
			"overdue":         overdue,
			"escalated":       esc,
			"recently_closed": rc,
		})
	})
}

func mountAudit(r *gin.RouterGroup, deps Router) {
	a := r.Group("/audit")
	a.GET("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermAuditRead}) {
			return
		}
		q := parseListQuery(c, 50)
		list, page, err := deps.Audit.List(c.Request.Context(), q)
		if err != nil {
			writeError(c, err)
			return
		}
		items := make([]gin.H, len(list))
		for i, e := range list {
			items[i] = gin.H{
				"id": e.ID, "tenant_id": e.TenantID, "actor_id": e.ActorID,
				"action": string(e.Action), "entity_type": e.EntityType,
				"entity_id": e.EntityID, "detail": e.Detail,
				"created_at": formatRFC3339(e.CreatedAt),
			}
		}
		c.JSON(http.StatusOK, gin.H{"items": items, "page": page.Page, "page_size": page.PageSize, "total": page.Total, "has_next": page.HasNext})
	})
	a.GET("/export", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermAuditExport}) {
			return
		}
		q := parseListQuery(c, 10000)
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", `attachment; filename="audit.csv"`)
		n, err := deps.Audit.ExportCSV(c.Request.Context(), q, c.Writer)
		if err != nil {
			writeError(c, err)
			return
		}
		_ = n
	})
	a.GET("/exceptions/:cycleId/export", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermAuditExport}) {
			return
		}
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", `attachment; filename="exceptions.csv"`)
		n, err := deps.Audit.ExportExceptionsCSV(c.Request.Context(), c.GetString("tenant_id"), c.Param("cycleId"), c.Writer)
		if err != nil {
			writeError(c, err)
			return
		}
		_ = n
	})
}

func mountUsers(r *gin.RouterGroup, deps Router) {
	u := r.Group("/users")
	u.POST("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermUserCreate}) {
			return
		}
		var req struct {
			Username     string `json:"username" binding:"required"`
			DisplayName  string `json:"display_name"`
			Email        string `json:"email"`
			Role         string `json:"role" binding:"required,oneof=operator assignee reviewer admin"`
			PasswordHash string `json:"password_hash"`
		}
		if !bindAndValidate(c, &req) {
			return
		}
		user, err := deps.Users.Create(c.Request.Context(), application.CreateUserInput{
			TenantID: c.GetString("tenant_id"), Username: req.Username, DisplayName: req.DisplayName,
			Email: req.Email, Role: domain.Role(req.Role), PasswordHash: req.PasswordHash,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeCreated(c, gin.H{
			"id": user.ID, "username": user.Username, "display_name": user.DisplayName,
			"email": user.Email, "role": string(user.Role),
		})
	})
	u.GET("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermWorkspaceRead}) {
			return
		}
		q := parseListQuery(c, 50)
		list, page, err := deps.Users.List(c.Request.Context(), q)
		if err != nil {
			writeError(c, err)
			return
		}
		items := make([]gin.H, len(list))
		for i, u := range list {
			items[i] = gin.H{
				"id": u.ID, "username": u.Username, "display_name": u.DisplayName,
				"email": u.Email, "role": string(u.Role),
			}
		}
		c.JSON(http.StatusOK, gin.H{"items": items, "page": page.Page, "page_size": page.PageSize, "total": page.Total, "has_next": page.HasNext})
	})
}
