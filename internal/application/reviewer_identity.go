package application
import "strings"
func normalizeReviewerID(id string) string { id=strings.TrimSpace(id); if id=="" {return ""}; return strings.ToLower(id) }
