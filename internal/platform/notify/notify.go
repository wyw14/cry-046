// Package notify is a local notification adapter. It records every
// notification in an in-memory ring buffer that tests can inspect; the
// buffer is bounded so the platform never holds unbounded data.
package notify

import (
	"context"
	"sync"
	"time"
)

// Message is the offline notification envelope.
type Message struct {
	ID          string
	Recipient   string
	Channel     string // "system" | "email-local" | "sms-local"
	Subject     string
	Body        string
	CreatedAt   time.Time
	DeliveredAt time.Time
}

// Adapter is the local notification adapter.
type Adapter struct {
	mu      sync.Mutex
	buffer  []Message
	maxSize int
	now     func() time.Time
}

// New constructs an Adapter with the given buffer capacity.
func New(maxSize int) *Adapter {
	if maxSize <= 0 {
		maxSize = 256
	}
	return &Adapter{
		buffer:  make([]Message, 0, maxSize),
		maxSize: maxSize,
		now:     time.Now,
	}
}

// Send records a notification in the local ring buffer.
func (a *Adapter) Send(ctx context.Context, msg Message) (Message, error) {
	if ctx.Err() != nil {
		return Message{}, ctx.Err()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = a.now()
	}
	msg.DeliveredAt = a.now()
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.buffer) >= a.maxSize {
		a.buffer = a.buffer[1:]
	}
	a.buffer = append(a.buffer, msg)
	return msg, nil
}

// Snapshot returns a copy of all buffered messages.
func (a *Adapter) Snapshot() []Message {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]Message, len(a.buffer))
	copy(out, a.buffer)
	return out
}

// Reset clears the buffer (used by tests and seed runs).
func (a *Adapter) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.buffer = a.buffer[:0]
}
