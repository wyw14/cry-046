package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-046/internal/application"
)

func (s *Server) createPalette(c *gin.Context) {
	var in application.CreatePaletteInput
	if c.ShouldBindJSON(&in) != nil {
		writeError(c, ErrBadJSON)
		return
	}
	p, err := s.Palettes.Create(c, in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(201, p)
}
func (s *Server) listPalettes(c *gin.Context) {
	p, err := s.Palettes.List(c, c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, gin.H{"items": p})
}
func (s *Server) submitPalette(c *gin.Context) {
	p, err := s.Palettes.Submit(c, c.Param("id"), c.GetString("actor_id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, p)
}
func (s *Server) approvePalette(c *gin.Context) {
	p, err := s.Palettes.Approve(c, c.Param("id"), c.GetString("actor_id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, p)
}
func (s *Server) deliverPalette(c *gin.Context) {
	p, err := s.Palettes.Deliver(c, c.Param("id"), c.GetString("actor_id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, p)
}

type deriveRequest struct{ ID, Name string }

func (s *Server) derivePalette(c *gin.Context) {
	var in deriveRequest
	if c.ShouldBindJSON(&in) != nil {
		writeError(c, ErrBadJSON)
		return
	}
	p, err := s.Palettes.Derive(c, c.Param("id"), in.ID, in.Name, c.GetString("actor_id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(201, p)
}
func (s *Server) diffPalette(c *gin.Context) {
	d, err := s.Palettes.Diff(c, c.Param("left"), c.Param("right"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, d)
}
