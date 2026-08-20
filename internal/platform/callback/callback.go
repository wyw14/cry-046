// Package callback is a local event callback adapter. It records every
// callback in memory and exposes an inspectable history so tests can
// assert that downstream systems would have been notified.
package callback

import (
	"context"
	"sync"
	"time"
)

// Event describes a single callback invocation.
type Event struct {
	ID        string
	Topic     string
	Payload   map[string]any
	CreatedAt time.Time
}

// Adapter is the local callback adapter.
type Adapter struct {
	mu      sync.Mutex
	events  []Event
	maxSize int
	now     func() time.Time
}

// New constructs an Adapter.
func New(maxSize int) *Adapter {
	if maxSize <= 0 {
		maxSize = 256
	}
	return &Adapter{
		events:  make([]Event, 0, maxSize),
		maxSize: maxSize,
		now:     time.Now,
	}
}

// Emit records a callback event.
func (a *Adapter) Emit(ctx context.Context, topic string, payload map[string]any) (Event, error) {
	if ctx.Err() != nil {
		return Event{}, ctx.Err()
	}
	ev := Event{
		ID:        "",
		Topic:     topic,
		Payload:   payload,
		CreatedAt: a.now(),
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.events) >= a.maxSize {
		a.events = a.events[1:]
	}
	a.events = append(a.events, ev)
	return ev, nil
}

// Snapshot returns a copy of all buffered events.
func (a *Adapter) Snapshot() []Event {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]Event, len(a.events))
	copy(out, a.events)
	return out
}

// Reset clears the events.
func (a *Adapter) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = a.events[:0]
}
