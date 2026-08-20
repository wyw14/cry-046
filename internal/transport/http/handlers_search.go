package httpapi

import "github.com/gin-gonic/gin"

func (s *Server) search(c *gin.Context) {
	out, err := s.Search.Search(c, c.Query("q"), queryInt(c, "page", 1), queryInt(c, "page_size", 20))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, out)
}
