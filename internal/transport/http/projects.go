package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
	"github.com/welfare/settlement-resolver/internal/service/authz"
	"github.com/welfare/settlement-resolver/internal/service/masking"
)

// ProjectDTO is the JSON shape for a project.
type ProjectDTO struct {
	ID           string            `json:"id"`
	TenantID     string            `json:"tenant_id"`
	Code         string            `json:"code"`
	Name         string            `json:"name"`
	Sponsor      string            `json:"sponsor"`
	AnnualBudget int64             `json:"annual_budget_cents"`
	StartYear    int               `json:"start_year"`
	EndYear      int               `json:"end_year"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
}

func toProjectDTO(p domain.Project) ProjectDTO {
	return ProjectDTO{
		ID: p.ID, TenantID: p.TenantID, Code: p.Code, Name: p.Name, Sponsor: p.Sponsor,
		AnnualBudget: p.AnnualBudget, StartYear: p.StartYear, EndYear: p.EndYear,
		Metadata: p.Metadata, CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type createProjectReq struct {
	Code         string            `json:"code" binding:"required"`
	Name         string            `json:"name" binding:"required"`
	Sponsor      string            `json:"sponsor" binding:"required"`
	AnnualBudget int64             `json:"annual_budget_cents" binding:"gte=0"`
	StartYear    int               `json:"start_year" binding:"required,gte=1900"`
	EndYear      int               `json:"end_year" binding:"required,gte=1900"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

func mountProjects(r *gin.RouterGroup, deps Router) {
	projects := r.Group("/projects")
	projects.POST("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermProjectCreate}) {
			return
		}
		var req createProjectReq
		if !bindAndValidate(c, &req) {
			return
		}
		p, err := deps.Projects.Create(c.Request.Context(), application.CreateProjectInput{
			TenantID:     c.GetString("tenant_id"),
			Code:         req.Code,
			Name:         req.Name,
			Sponsor:      req.Sponsor,
			AnnualBudget: req.AnnualBudget,
			StartYear:    req.StartYear,
			EndYear:      req.EndYear,
			Metadata:     req.Metadata,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeCreated(c, toProjectDTO(p))
	})
	projects.GET("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermProjectRead}) {
			return
		}
		q := parseListQuery(c, 20)
		list, page, err := deps.Projects.List(c.Request.Context(), q)
		if err != nil {
			writeError(c, err)
			return
		}
		out := make([]ProjectDTO, len(list))
		for i, p := range list {
			out[i] = toProjectDTO(p)
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "page": page.Page, "page_size": page.PageSize, "total": page.Total, "has_next": page.HasNext})
	})
	projects.GET("/:id", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermProjectRead}) {
			return
		}
		p, err := deps.Projects.Get(c.Request.Context(), c.GetString("tenant_id"), c.Param("id"))
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toProjectDTO(p))
	})
	projects.PATCH("/:id", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermProjectCreate}) {
			return
		}
		var req createProjectReq
		if !bindAndValidate(c, &req) {
			return
		}
		p, err := deps.Projects.Get(c.Request.Context(), c.GetString("tenant_id"), c.Param("id"))
		if err != nil {
			writeError(c, err)
			return
		}
		p.Name = req.Name
		p.Sponsor = req.Sponsor
		p.AnnualBudget = req.AnnualBudget
		p.StartYear = req.StartYear
		p.EndYear = req.EndYear
		if req.Metadata != nil {
			p.Metadata = req.Metadata
		}
		updated, err := deps.Projects.Update(c.Request.Context(), p)
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toProjectDTO(updated))
	})
}

// PartyDTO is the JSON shape for a party.
type PartyDTO struct {
	ID        string            `json:"id"`
	TenantID  string            `json:"tenant_id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Contact   string            `json:"contact"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

func toPartyDTO(p domain.Party) PartyDTO {
	return PartyDTO{
		ID: p.ID, TenantID: p.TenantID, Name: p.Name, Type: string(p.Type),
		Contact: masking.MaskString(p.Contact), Metadata: p.Metadata,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type createPartyReq struct {
	Name     string            `json:"name" binding:"required"`
	Type     string            `json:"type" binding:"required,oneof=donor implementer beneficiary intermediary"`
	Contact  string            `json:"contact" binding:"required"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func mountParties(r *gin.RouterGroup, deps Router) {
	parties := r.Group("/parties")
	parties.POST("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermPartyCreate}) {
			return
		}
		var req createPartyReq
		if !bindAndValidate(c, &req) {
			return
		}
		p, err := deps.Parties.Create(c.Request.Context(), application.CreatePartyInput{
			TenantID: c.GetString("tenant_id"),
			Name:     req.Name, Type: domain.PartyType(req.Type), Contact: req.Contact, Metadata: req.Metadata,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeCreated(c, toPartyDTO(p))
	})
	parties.GET("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermProjectRead}) {
			return
		}
		q := parseListQuery(c, 20)
		list, page, err := deps.Parties.List(c.Request.Context(), q)
		if err != nil {
			writeError(c, err)
			return
		}
		out := make([]PartyDTO, len(list))
		for i, p := range list {
			out[i] = toPartyDTO(p)
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "page": page.Page, "page_size": page.PageSize, "total": page.Total, "has_next": page.HasNext})
	})
	parties.GET("/:id", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermProjectRead}) {
			return
		}
		p, err := deps.Parties.Get(c.Request.Context(), c.GetString("tenant_id"), c.Param("id"))
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toPartyDTO(p))
	})
}

// BatchDTO is the JSON shape for a funding batch.
type BatchDTO struct {
	ID                  string            `json:"id"`
	TenantID            string            `json:"tenant_id"`
	ProjectID           string            `json:"project_id"`
	Code                string            `json:"code"`
	DonorPartyID        string            `json:"donor_party_id"`
	ImplementerPartyID  string            `json:"implementer_party_id"`
	IntermediaryPartyID string            `json:"intermediary_party_id"`
	TotalAmountCents    int64             `json:"total_amount_cents"`
	Currency            string            `json:"currency"`
	DisbursedAt         string            `json:"disbursed_at"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	CreatedAt           string            `json:"created_at"`
	UpdatedAt           string            `json:"updated_at"`
}

func toBatchDTO(b domain.FundingBatch) BatchDTO {
	return BatchDTO{
		ID: b.ID, TenantID: b.TenantID, ProjectID: b.ProjectID, Code: b.Code,
		DonorPartyID: b.DonorPartyID, ImplementerPartyID: b.ImplementerPartyID,
		IntermediaryPartyID: b.IntermediaryPartyID, TotalAmountCents: b.TotalAmount,
		Currency: b.Currency, DisbursedAt: b.DisbursedAt.Format("2006-01-02T15:04:05Z07:00"),
		Metadata:  b.Metadata,
		CreatedAt: b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type createBatchReq struct {
	ProjectID           string            `json:"project_id" binding:"required"`
	Code                string            `json:"code" binding:"required"`
	DonorPartyID        string            `json:"donor_party_id" binding:"required"`
	ImplementerPartyID  string            `json:"implementer_party_id" binding:"required"`
	IntermediaryPartyID string            `json:"intermediary_party_id,omitempty"`
	TotalAmountCents    int64             `json:"total_amount_cents" binding:"required,gt=0"`
	Currency            string            `json:"currency" binding:"required"`
	DisbursedAt         string            `json:"disbursed_at" binding:"required"`
	Metadata            map[string]string `json:"metadata,omitempty"`
}

func mountBatches(r *gin.RouterGroup, deps Router) {
	batches := r.Group("/batches")
	batches.POST("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermBatchCreate}) {
			return
		}
		var req createBatchReq
		if !bindAndValidate(c, &req) {
			return
		}
		disbursedAt, err := parseRFC3339(req.DisbursedAt)
		if err != nil {
			writeError(c, err)
			return
		}
		b, err := deps.Batches.Create(c.Request.Context(), application.CreateBatchInput{
			TenantID: c.GetString("tenant_id"), ProjectID: req.ProjectID, Code: req.Code,
			DonorPartyID: req.DonorPartyID, ImplementerPartyID: req.ImplementerPartyID,
			IntermediaryPartyID: req.IntermediaryPartyID, TotalAmount: req.TotalAmountCents,
			Currency: req.Currency, DisbursedAt: disbursedAt, Metadata: req.Metadata,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeCreated(c, toBatchDTO(b))
	})
	batches.GET("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermProjectRead}) {
			return
		}
		q := parseListQuery(c, 20)
		list, page, err := deps.Batches.List(c.Request.Context(), q)
		if err != nil {
			writeError(c, err)
			return
		}
		out := make([]BatchDTO, len(list))
		for i, b := range list {
			out[i] = toBatchDTO(b)
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "page": page.Page, "page_size": page.PageSize, "total": page.Total, "has_next": page.HasNext})
	})
	batches.GET("/:id", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermProjectRead}) {
			return
		}
		b, err := deps.Batches.Get(c.Request.Context(), c.GetString("tenant_id"), c.Param("id"))
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toBatchDTO(b))
	})
}

// CycleDTO is the JSON shape for a settlement cycle.
type CycleDTO struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenant_id"`
	ProjectID      string `json:"project_id"`
	FundingBatchID string `json:"funding_batch_id"`
	Year           int    `json:"year"`
	Quarter        int    `json:"quarter"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	ClosedAt       string `json:"closed_at,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func toCycleDTO(c domain.SettlementCycle) CycleDTO {
	out := CycleDTO{
		ID: c.ID, TenantID: c.TenantID, ProjectID: c.ProjectID,
		FundingBatchID: c.FundingBatchID, Year: c.Year, Quarter: c.Quarter,
		StartDate: c.StartDate.Format("2006-01-02T15:04:05Z07:00"),
		EndDate:   c.EndDate.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if !c.ClosedAt.IsZero() {
		out.ClosedAt = c.ClosedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return out
}

type createCycleReq struct {
	ProjectID      string `json:"project_id" binding:"required"`
	FundingBatchID string `json:"funding_batch_id" binding:"required"`
	Year           int    `json:"year" binding:"required,gte=1900"`
	Quarter        int    `json:"quarter" binding:"required,oneof=1 2 3 4"`
	StartDate      string `json:"start_date" binding:"required"`
	EndDate        string `json:"end_date" binding:"required"`
}

func mountCycles(r *gin.RouterGroup, deps Router) {
	cycles := r.Group("/cycles")
	cycles.POST("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermCycleCreate}) {
			return
		}
		var req createCycleReq
		if !bindAndValidate(c, &req) {
			return
		}
		start, err := parseRFC3339(req.StartDate)
		if err != nil {
			writeError(c, err)
			return
		}
		end, err := parseRFC3339(req.EndDate)
		if err != nil {
			writeError(c, err)
			return
		}
		cc, err := deps.Cycles.Create(c.Request.Context(), application.CreateCycleInput{
			TenantID: c.GetString("tenant_id"), ProjectID: req.ProjectID,
			FundingBatchID: req.FundingBatchID, Year: req.Year, Quarter: req.Quarter,
			StartDate: start, EndDate: end,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		writeCreated(c, toCycleDTO(cc))
	})
	cycles.GET("", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermProjectRead}) {
			return
		}
		q := parseListQuery(c, 20)
		list, page, err := deps.Cycles.List(c.Request.Context(), q)
		if err != nil {
			writeError(c, err)
			return
		}
		out := make([]CycleDTO, len(list))
		for i, cc := range list {
			out[i] = toCycleDTO(cc)
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "page": page.Page, "page_size": page.PageSize, "total": page.Total, "has_next": page.HasNext})
	})
	cycles.GET("/:id", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermProjectRead}) {
			return
		}
		cc, err := deps.Cycles.Get(c.Request.Context(), c.GetString("tenant_id"), c.Param("id"))
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toCycleDTO(cc))
	})
	cycles.POST("/:id/close", func(c *gin.Context) {
		if !requireRole(c, []authz.Permission{authz.PermCycleCreate}) {
			return
		}
		cc, err := deps.Cycles.Close(c.Request.Context(), c.GetString("tenant_id"), c.Param("id"))
		if err != nil {
			writeError(c, err)
			return
		}
		writeOK(c, toCycleDTO(cc))
	})
}
