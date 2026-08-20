package domain

import "time"

type Favorite struct {
	UserID, ProjectID string
	CreatedAt         time.Time `json:"created_at"`
}
type RecentVisit struct {
	UserID, ProjectID string
	VisitedAt         time.Time `json:"visited_at"`
}
type Todo struct {
	ID, UserID, ProjectID, Kind, Summary string
	DueAt                                time.Time
	Done                                 bool
	CreatedAt                            time.Time `json:"created_at"`
}
type AccessLevel string

const (
	AccessViewer AccessLevel = "viewer"
	AccessEditor AccessLevel = "editor"
	AccessOwner  AccessLevel = "owner"
)

func (l AccessLevel) CanWrite() bool   { return l == AccessEditor || l == AccessOwner }
func (l AccessLevel) CanApprove() bool { return l == AccessOwner }
