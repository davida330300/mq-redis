package api

import (
	"context"
	"encoding/json"
)

type Store interface {
	GetJobIDByIdempotencyKey(ctx context.Context, key string) (jobID string, found bool, err error)
	CreateJob(ctx context.Context, key, jobID string, payload json.RawMessage) error
}

type Producer interface {
	Publish(ctx context.Context, jobID string, payload json.RawMessage) error
}
