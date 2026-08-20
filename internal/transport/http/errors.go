package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-046/internal/repository/memory"
	"github.com/wyw14/cry-046/internal/service"
)

func writeError(c *gin.Context, err error) {
	code, status := "INVALID", http.StatusBadRequest
	switch {
	case errors.Is(err, memory.ErrNotFound), errors.Is(err, service.ErrNotFound):
		code, status = "NOT_FOUND", 404
	case errors.Is(err, service.ErrForbidden):
		code, status = "FORBIDDEN", 403
	case errors.Is(err, memory.ErrConflict):
		code, status = "STALE_VERSION", 409
	case errors.Is(err, memory.ErrDuplicate):
		code, status = "DUPLICATE", 409
	}
	c.JSON(status, gin.H{"code": code, "message": err.Error(), "request_id": c.GetString("request_id")})
}
