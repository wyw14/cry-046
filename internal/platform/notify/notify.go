package notify

import (
	"context"
	"sync"
)

type Message struct{ Recipient, Body string }
type Inbox struct {
	mu       sync.RWMutex
	messages map[string][]Message
}

func New() *Inbox { return &Inbox{messages: map[string][]Message{}} }
func (i *Inbox) Notify(_ context.Context, recipient, body string) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.messages[recipient] = append(i.messages[recipient], Message{Recipient: recipient, Body: body})
	return nil
}
func (i *Inbox) List(recipient string) []Message {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return append([]Message(nil), i.messages[recipient]...)
}
