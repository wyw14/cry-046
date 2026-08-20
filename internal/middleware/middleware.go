// Package middleware contains the Gin middlewares used by the
// transport layer: request id, structured logging, panic recovery,
// CORS, security headers, audit-actor injection and graceful shutdown.
package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// RequestID middleware stamps every request with a unique id and
// exposes it via the X-Request-Id response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set("request_id", rid)
		c.Header("X-Request-Id", rid)
		c.Next()
	}
}

// Logger logs each request as a structured zap log entry.
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		logger.Info("http",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.Duration("latency", latency),
			zap.String("request_id", c.GetString("request_id")),
			zap.String("remote", c.ClientIP()),
		)
	}
}

// Recover catches panics, logs them and returns a 500 envelope.
func Recover(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic",
					zap.Any("recover", r),
					zap.String("request_id", c.GetString("request_id")),
					zap.ByteString("stack", debug.Stack()),
				)
				env := ErrorEnvelope{
					Code:      domain.CodeUnknown,
					Message:   "internal server error",
					RequestID: c.GetString("request_id"),
				}
				c.AbortWithStatusJSON(http.StatusInternalServerError, env)
			}
		}()
		c.Next()
	}
}

// ErrorEnvelope mirrors the transport ErrorEnvelope so this package
// does not import the transport layer.
type ErrorEnvelope struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Fields    []FieldError `json:"fields,omitempty"`
	RequestID string       `json:"request_id"`
}

// FieldError is a single field-level validation error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// CORS allows the configured origins. Defaults to localhost:5173.
func CORS(allowed []string) gin.HandlerFunc {
	allow := strings.Join(allowed, ", ")
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && contains(allowed, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Request-Id,X-Tenant-ID")
			c.Header("Access-Control-Expose-Headers", "X-Request-Id")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		_ = allow
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func contains(in []string, needle string) bool {
	for _, v := range in {
		if v == needle {
			return true
		}
	}
	return false
}

// SecurityHeaders sets conservative security headers on every response.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-XSS-Protection", "1; mode=block")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		h.Set("Cache-Control", "no-store")
		c.Next()
	}
}

// TenantResolver extracts the tenant id from the X-Tenant-ID header
// or the sub-domain. For the demo we only support the header.
func TenantResolver(defaultTenant string) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := c.GetHeader("X-Tenant-ID")
		if t == "" {
			t = defaultTenant
		}
		if t == "" {
			t = "default"
		}
		c.Set("tenant_id", t)
		c.Next()
	}
}

// Actor is the authenticated operator that performed the request.
type Actor struct {
	UserID   string
	Username string
	Role     domain.Role
	TenantID string
}

// ActorFromContext returns the actor stamped by Auth middleware.
func ActorFromContext(c *gin.Context) (Actor, bool) {
	v, ok := c.Get("actor")
	if !ok {
		return Actor{}, false
	}
	a, ok := v.(Actor)
	return a, ok
}

// Auth is a local-only auth middleware. It accepts a Bearer token that
// is just the user id (for the offline demo). Real deployments would
// replace this with a session/JWT check.
func Auth(users map[string]Actor) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorEnvelope{
				Code:      domain.CodeUnauthenticated,
				Message:   "missing or invalid authorization header",
				RequestID: c.GetString("request_id"),
			})
			return
		}
		token := strings.TrimPrefix(authHeader, prefix)
		a, ok := users[token]
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorEnvelope{
				Code:      domain.CodeUnauthenticated,
				Message:   "unknown token",
				RequestID: c.GetString("request_id"),
			})
			return
		}
		c.Set("actor", a)
		c.Next()
	}
}

// RequirePermission aborts with 403 if the actor lacks the permission.
func RequirePermission(perm string, allowed func(role domain.Role) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		a, ok := ActorFromContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorEnvelope{
				Code:      domain.CodeUnauthenticated,
				Message:   "actor required",
				RequestID: c.GetString("request_id"),
			})
			return
		}
		if !allowed(a.Role) {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorEnvelope{
				Code:      domain.CodePermissionDenied,
				Message:   "permission denied for " + perm,
				RequestID: c.GetString("request_id"),
			})
			return
		}
		c.Next()
	}
}
