package memory

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
)

var ErrAlreadyExists = errors.New("idempotency key already exists")

// Store is an in-memory implementation of the API Store interface.
type Store struct {
	mu       sync.RWMutex
	byKey    map[string]string
	payloads map[string]json.RawMessage
}

func New() *Store {
	return &Store{
		byKey:    make(map[string]string),
		payloads: make(map[string]json.RawMessage),
	}
}

func (s *Store) GetJobIDByIdempotencyKey(ctx context.Context, key string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	jobID, ok := s.byKey[key]
	return jobID, ok, nil
}

func (s *Store) CreateJob(ctx context.Context, key, jobID string, payload json.RawMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byKey[key]; exists {
		return ErrAlreadyExists
	}
	s.byKey[key] = jobID
	s.payloads[jobID] = payload
	return nil
}
