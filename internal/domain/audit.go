package domain

import "time"

type AuditEvent struct {
	ID        string            `json:"id"`
	ActorID   string            `json:"actor_id"`
	Action    string            `json:"action"`
	Entity    string            `json:"entity"`
	EntityID  string            `json:"entity_id"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}
