package domain

import (
	"errors"
	"strings"
	"time"
)

type Confidentiality string

const (
	ConfPublic Confidentiality = "public"
	ConfTeam   Confidentiality = "team"
	ConfSecret Confidentiality = "secret"
)

type ProjectState string

const (
	ProjectActive   ProjectState = "active"
	ProjectArchived ProjectState = "archived"
)

type Project struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Customer     string          `json:"customer"`
	Scene        string          `json:"scene"`
	Series       string          `json:"series"`
	Tags         []string        `json:"tags"`
	Confidential Confidentiality `json:"confidentiality"`
	State        ProjectState    `json:"state"`
	OwnerID      string          `json:"owner_id"`
	CreatedAt    time.Time       `json:"created_at"`
	ArchivedAt   *time.Time      `json:"archived_at,omitempty"`
	Version      int             `json:"version"`
}

func NewProject(id, name, owner string) (Project, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(owner) == "" {
		return Project{}, errors.New("project id, name and owner are required")
	}
	return Project{ID: id, Name: name, OwnerID: owner, State: ProjectActive, Confidential: ConfTeam, CreatedAt: time.Now().UTC(), Version: 1}, nil
}

func (p Project) CanEdit() bool { return p.State == ProjectActive }

func (p *Project) Archive(now time.Time) error {
	if p.State == ProjectArchived {
		return errors.New("project already archived")
	}
	p.State = ProjectArchived
	p.ArchivedAt = &now
	p.Version = p.Version + 1
	return nil
}
