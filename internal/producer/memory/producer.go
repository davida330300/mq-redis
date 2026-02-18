package memory

import (
	"context"
	"encoding/json"
	"sync"
)

// Producer is an in-memory implementation of the API Producer interface.
type Producer struct {
	mu        sync.Mutex
	published []Message
}

type Message struct {
	JobID   string
	Payload json.RawMessage
}

func New() *Producer {
	return &Producer{}
}

func (p *Producer) Publish(ctx context.Context, jobID string, payload json.RawMessage) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.published = append(p.published, Message{JobID: jobID, Payload: payload})
	return nil
}

func (p *Producer) Published() []Message {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]Message, len(p.published))
	copy(out, p.published)
	return out
}
