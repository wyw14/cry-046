package audit

import (
	"context"
	"sync"

	"github.com/wyw14/cry-046/internal/domain"
)

type Log struct {
	mu     sync.RWMutex
	events []domain.AuditEvent
}

func New() *Log { return &Log{} }
func (l *Log) Append(_ context.Context, e domain.AuditEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, e)
	return nil
}
func (l *Log) List(_ context.Context, entity string, limit int) ([]domain.AuditEvent, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := []domain.AuditEvent{}
	for i := len(l.events) - 1; i >= 0 && len(out) < limit; i-- {
		if entity == "" || l.events[i].EntityID == entity {
			out = append(out, l.events[i])
		}
	}
	return out, nil
}
