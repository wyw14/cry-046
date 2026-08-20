package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-046/internal/application"
)

func (s *Server) requestDelivery(c *gin.Context) {
	var in application.CreateDeliveryInput
	if c.ShouldBindJSON(&in) != nil {
		writeError(c, ErrBadJSON)
		return
	}
	in.Applicant = c.GetString("actor_id")
	d, err := s.Deliveries.Request(c, in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(201, d)
}
func (s *Server) approveDelivery(c *gin.Context) {
	d, err := s.Deliveries.Approve(c, c.Param("id"), c.GetString("actor_id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, d)
}
func (s *Server) downloadDelivery(c *gin.Context) {
	path, err := s.Deliveries.Download(c, c.Param("id"), c.GetString("actor_id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, gin.H{"path": path})
}
func (s *Server) listDeliveries(c *gin.Context) {
	d, err := s.Deliveries.List(c, c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, gin.H{"items": d})
}
