package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = fmt.Sprintf("req-%d", time.Now().UnixNano())
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Error("panic recovered", zap.Any("error", err))
		c.AbortWithStatusJSON(500, gin.H{"code": "INTERNAL", "message": "internal server error", "request_id": c.GetString("request_id")})
	})
}
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Next()
	}
}
func Actor() gin.HandlerFunc {
	return func(c *gin.Context) {
		actor := c.GetHeader("X-Actor-ID")
		if actor == "" {
			actor = "demo-owner"
		}
		c.Set("actor_id", actor)
		c.Next()
	}
}
