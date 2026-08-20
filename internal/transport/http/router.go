package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-046/internal/middleware"
	"github.com/wyw14/cry-046/internal/service"
	"go.uber.org/zap"
)

type Server struct {
	Projects   *service.ProjectService
	Assets     *service.AssetService
	Palettes   *service.PaletteService
	Deliveries *service.DeliveryService
	Reviews    *service.ReviewService
	Search     *service.SearchService
	Workspace  *service.Workspace
	log        *zap.Logger
}

func NewServer(p *service.ProjectService, a *service.AssetService, ps *service.PaletteService, d *service.DeliveryService, r *service.ReviewService, s *service.SearchService, w *service.Workspace, log *zap.Logger) *Server {
	return &Server{Projects: p, Assets: a, Palettes: ps, Deliveries: d, Reviews: r, Search: s, Workspace: w, log: log}
}
func (s *Server) Router() *gin.Engine {
	g := gin.New()
	g.Use(middleware.RequestID(), middleware.Actor(), middleware.Security(), middleware.Recovery(s.log))
	g.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	g.GET("/readyz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })
	api := g.Group("/api/v1")
	api.POST("/projects", s.createProject)
	api.GET("/projects", s.listProjects)
	api.POST("/projects/:id/archive", s.archiveProject)
	api.POST("/assets", s.createAsset)
	api.GET("/projects/:id/assets", s.listAssets)
	api.POST("/palettes", s.createPalette)
	api.GET("/projects/:id/palettes", s.listPalettes)
	api.POST("/palettes/:id/submit", s.submitPalette)
	api.POST("/palettes/:id/approve", s.approvePalette)
	api.POST("/palettes/:id/deliver", s.deliverPalette)
	api.POST("/palettes/:id/derive", s.derivePalette)
	api.GET("/palettes/:left/diff/:right", s.diffPalette)
	api.POST("/deliveries", s.requestDelivery)
	api.POST("/deliveries/:id/approve", s.approveDelivery)
	api.GET("/deliveries/:id/download", s.downloadDelivery)
	api.GET("/projects/:id/deliveries", s.listDeliveries)
	api.GET("/search", s.search)
	api.POST("/projects/:id/favorite", s.toggleFavorite)
	api.POST("/projects/:id/visit", s.visitProject)
	api.GET("/favorites", s.favorites)
	api.GET("/recent", s.recent)
	api.GET("/todos", s.todos)
	api.POST("/todos/:id/complete", s.completeTodo)
	return g
}
