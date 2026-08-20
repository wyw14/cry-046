package httpapi

import "github.com/gin-gonic/gin"

func (s *Server) toggleFavorite(c *gin.Context) {
	ok, err := s.Workspace.ToggleFavorite(c, c.GetString("actor_id"), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, gin.H{"favorite": ok})
}
func (s *Server) visitProject(c *gin.Context) {
	if err := s.Workspace.RecordVisit(c, c.GetString("actor_id"), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(204)
}
func (s *Server) favorites(c *gin.Context) {
	items, err := s.Workspace.ListFavorites(c, c.GetString("actor_id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, gin.H{"items": items})
}
func (s *Server) recent(c *gin.Context) {
	items, err := s.Workspace.ListRecent(c, c.GetString("actor_id"), queryInt(c, "limit", 20))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, gin.H{"items": items})
}
func (s *Server) todos(c *gin.Context) {
	items, err := s.Workspace.ListTodos(c, c.GetString("actor_id"), c.Query("include_done") == "true")
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, gin.H{"items": items})
}
func (s *Server) completeTodo(c *gin.Context) {
	if err := s.Workspace.CompleteTodo(c, c.GetString("actor_id"), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(204)
}
