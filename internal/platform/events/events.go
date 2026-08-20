package events

import (
	"context"
	"sync"
)

type Event struct {
	Name    string
	Payload map[string]any
}
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Event
}

func New() *Bus { return &Bus{subscribers: map[string][]chan Event{}} }
func (b *Bus) Subscribe(name string, buffer int) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan Event, buffer)
	b.subscribers[name] = append(b.subscribers[name], ch)
	return ch
}
func (b *Bus) Publish(_ context.Context, e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers[e.Name] {
		select {
		case ch <- e:
		default:
		}
	}
}
