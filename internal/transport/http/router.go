package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
	"github.com/welfare/settlement-resolver/internal/service/authz"
)

// Router bundles all application services the transport layer needs.
// It is constructed once at startup and shared across handlers.
type Router struct {
	Projects   *application.ProjectsApp
	Parties    *application.PartiesApp
	Batches    *application.BatchesApp
	Cycles     *application.CyclesApp
	Rules      *application.RulesApp
	Imports    *application.ImportsApp
	Exceptions *application.ExceptionsApp
	Summary    *application.SummaryApp
	Workspace  *application.WorkspaceApp
	Audit      *application.AuditApp
	Users      *application.UsersApp
	Evaluate   *application.EvaluateApp
}

// Actor is the operator stamped by the Auth middleware. It is exported
// so cmd/server can construct demo actors for the offline token map.
type Actor struct {
	UserID   string
	Username string
	Role     domain.Role
	TenantID string
}

// New returns a Gin engine with all routes mounted under /api/v1.
func New(r Router, opts ...Option) *gin.Engine {
	cfg := routerConfig{
		defaultTenant: "default",
		actors:        map[string]Actor{},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.GET("/healthz", healthz)
	engine.GET("/readyz", readyz(cfg))

	api := engine.Group("/api/v1")

	api.Use(
		requestID(),
		tenantResolver(cfg.defaultTenant),
	)
	api.Use(actorMiddleware(cfg.actors))
	api.Use(permissionMiddleware())

	mountProjects(api, r)
	mountParties(api, r)
	mountBatches(api, r)
	mountCycles(api, r)
	mountRules(api, r)
	mountEntries(api, r)
	mountExceptions(api, r)
	mountSummary(api, r)
	mountWorkspace(api, r)
	mountAudit(api, r)
	mountUsers(api, r)

	return engine
}

type routerConfig struct {
	defaultTenant string
	actors        map[string]Actor
}

// Option customises the Router.
type Option func(*routerConfig)

// WithDefaultTenant sets the default tenant id.
func WithDefaultTenant(t string) Option { return func(c *routerConfig) { c.defaultTenant = t } }

// WithActor registers an actor identified by a Bearer token.
func WithActor(token string, a Actor) Option {
	return func(c *routerConfig) { c.actors[token] = a }
}

// healthz returns 200 if the process is alive.
func healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// readyz returns 200 if the process is ready to serve traffic.
func readyz(_ routerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}

// requestID wraps the request id middleware to avoid importing the
// middleware package name in this file.
func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" {
			rid = domain.NewID()
		}
		c.Set("request_id", rid)
		c.Header("X-Request-Id", rid)
		c.Next()
	}
}

func tenantResolver(defaultTenant string) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := c.GetHeader("X-Tenant-ID")
		if t == "" {
			t = defaultTenant
		}
		c.Set("tenant_id", t)
		c.Next()
	}
}

func actorMiddleware(actors map[string]Actor) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Get("actor"); ok {
			c.Next()
			return
		}
		authHeader := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			// allow unauthenticated requests in development mode
			c.Next()
			return
		}
		token := strings.TrimPrefix(authHeader, prefix)
		a, ok := actors[token]
		if !ok {
			c.Next()
			return
		}
		c.Set("actor", a)
		c.Next()
	}
}

func permissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// actorFromContext returns the actor stamped by actorMiddleware.
func actorFromContext(c *gin.Context) (Actor, bool) {
	v, ok := c.Get("actor")
	if !ok {
		return Actor{}, false
	}
	a, ok := v.(Actor)
	return a, ok
}

// requireRole aborts with 403 if the actor lacks any of the allowed roles.
func requireRole(c *gin.Context, perms []authz.Permission) bool {
	a, ok := actorFromContext(c)
	if !ok {
		writeError(c, domain.NewErr(domain.CodeUnauthenticated, "actor required"))
		return false
	}
	for _, p := range perms {
		allowed := authz.AllowedRoles(p)
		for _, r := range allowed {
			if a.Role == r {
				return true
			}
		}
	}
	writeError(c, domain.NewErrf(domain.CodePermissionDenied, "permission denied for %v", perms))
	return false
}

// parseListQuery extracts the common list-query parameters from the
// gin context. Filters are taken from any query parameter that starts
// with "filter." and the value is a non-empty string. OrderBy defaults
// to "id"; OrderDesc is set when order=desc.
func parseListQuery(c *gin.Context, defaultPageSize int) application.ListQuery {
	q := application.ListQuery{
		TenantID:  c.GetString("tenant_id"),
		Page:      parseIntDefault(c.Query("page"), 1),
		PageSize:  parseIntDefault(c.Query("page_size"), defaultPageSize),
		OrderBy:   c.DefaultQuery("order_by", "id"),
		OrderDesc: c.DefaultQuery("order", "asc") == "desc",
		Filters:   map[string]string{},
	}
	for k, v := range c.Request.URL.Query() {
		if !strings.HasPrefix(k, "filter.") || len(v) == 0 || v[0] == "" {
			continue
		}
		q.Filters[strings.TrimPrefix(k, "filter.")] = v[0]
	}
	return q
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
