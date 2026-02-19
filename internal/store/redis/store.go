package redisstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"

	"mq-redis/internal/rediskeys"
	"mq-redis/internal/state"
	"mq-redis/internal/store"
)

type Store struct {
	client *redis.Client
}

func New(opts *redis.Options) *Store {
	return &Store{client: redis.NewClient(opts)}
}

func NewWithClient(client *redis.Client) *Store {
	return &Store{client: client}
}

func (s *Store) Close() error {
	return s.client.Close()
}

func (s *Store) GetJobIDByIdempotencyKey(ctx context.Context, key string) (string, bool, error) {
	val, err := s.client.Get(ctx, rediskeys.IdempotencyKey(key)).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, store.ErrStoreUnavailable
	}
	return val, true, nil
}

func (s *Store) CreateJob(ctx context.Context, key, jobID string, payload json.RawMessage) error {
	idemKey := rediskeys.IdempotencyKey(key)
	jobKey := rediskeys.JobKey(jobID)
	jobDataKey := rediskeys.JobDataKey(jobID)

	err := s.client.Watch(ctx, func(tx *redis.Tx) error {
		exists, err := tx.Exists(ctx, idemKey).Result()
		if err != nil {
			return err
		}
		if exists > 0 {
			return store.ErrAlreadyExists
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, idemKey, jobID, rediskeys.DedupeTTL)
			pipe.Set(ctx, jobKey, string(state.Queued), rediskeys.JobStatusTTL)
			pipe.Set(ctx, jobDataKey, []byte(payload), rediskeys.JobDataTTL)
			return nil
		})
		return err
	}, idemKey)
	if err == nil {
		return nil
	}
	if errors.Is(err, store.ErrAlreadyExists) || errors.Is(err, redis.TxFailedErr) {
		return store.ErrAlreadyExists
	}
	return fmt.Errorf("%w: %v", store.ErrStoreUnavailable, err)
}
