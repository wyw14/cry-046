package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-046/internal/application"
)

func (s *Server) createAsset(c *gin.Context) {
	var in application.CreateAssetInput
	if c.ShouldBindJSON(&in) != nil {
		writeError(c, ErrBadJSON)
		return
	}
	a, err := s.Assets.Create(c, in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(201, a)
}
func (s *Server) listAssets(c *gin.Context) {
	a, err := s.Assets.List(c, c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, gin.H{"items": a})
}
