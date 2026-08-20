package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/domain"
	"net/http"
)

func (s *Server) createProject(c *gin.Context) {
	var in application.CreateProjectInput
	if c.ShouldBindJSON(&in) != nil {
		writeError(c, ErrBadJSON)
		return
	}
	in.OwnerID = c.GetString("actor_id")
	if in.Confidentiality == "" {
		in.Confidentiality = domain.ConfTeam
	}
	p, err := s.Projects.Create(c, in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}
func (s *Server) listProjects(c *gin.Context) {
	items, total, err := s.Projects.List(c, c.Query("q"), queryInt(c, "page", 1), queryInt(c, "page_size", 20))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, gin.H{"items": items, "total": total})
}
func (s *Server) archiveProject(c *gin.Context) {
	p, err := s.Projects.Archive(c, c.Param("id"), c.GetString("actor_id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, p)
}
