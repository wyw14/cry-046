package httpapi

import (
	"errors"
	"github.com/gin-gonic/gin"
	"strconv"
)

var ErrBadJSON = errors.New("invalid json body")

func queryInt(c *gin.Context, key string, def int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil || v < 1 {
		return def
	}
	return v
}
