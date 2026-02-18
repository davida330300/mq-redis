package redisstore

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"mq-redis/internal/rediskeys"
	storeerr "mq-redis/internal/store"
)

func newTestStore(t *testing.T) (*Store, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis run: %v", err)
	}
	store := New(&redis.Options{Addr: mr.Addr(), DialTimeout: 500 * time.Millisecond})
	ctx := context.Background()
	var pingErr error
	for i := 0; i < 5; i++ {
		if err := store.client.Ping(ctx).Err(); err == nil {
			pingErr = nil
			break
		} else {
			pingErr = err
			time.Sleep(10 * time.Millisecond)
		}
	}
	if pingErr != nil {
		t.Fatalf("redis ping failed: %v", pingErr)
	}
	return store, mr
}

func TestStore_CreateAndGet(t *testing.T) {
	store, mr := newTestStore(t)
	defer mr.Close()
	defer store.Close()

	payload := json.RawMessage(`{"a":1}`)
	if err := store.CreateJob(context.Background(), "key1", "job1", payload); err != nil {
		t.Fatalf("CreateJob error: %v", err)
	}

	jobID, found, err := store.GetJobIDByIdempotencyKey(context.Background(), "key1")
	if err != nil {
		t.Fatalf("GetJobIDByIdempotencyKey error: %v", err)
	}
	if !found || jobID != "job1" {
		t.Fatalf("expected job1, got %q (found=%v)", jobID, found)
	}

	if got, err := mr.Get(rediskeys.JobKey("job1")); err != nil || got != "queued" {
		t.Fatalf("job status not set: %q err=%v", got, err)
	}
	if got, err := mr.Get(rediskeys.JobDataKey("job1")); err != nil || got != string(payload) {
		t.Fatalf("job data not set: %q err=%v", got, err)
	}

	if ttl := mr.TTL(rediskeys.IdempotencyKey("key1")); ttl <= 0*time.Second {
		t.Fatalf("expected dedupe TTL to be set, ttl=%v", ttl)
	}
	if ttl := mr.TTL(rediskeys.JobKey("job1")); ttl <= 0*time.Second {
		t.Fatalf("expected job TTL to be set, ttl=%v", ttl)
	}
	if ttl := mr.TTL(rediskeys.JobDataKey("job1")); ttl <= 0*time.Second {
		t.Fatalf("expected job data TTL to be set, ttl=%v", ttl)
	}
}

func TestStore_CreateDuplicate(t *testing.T) {
	store, mr := newTestStore(t)
	defer mr.Close()
	defer store.Close()

	payload := json.RawMessage(`{"a":1}`)
	if err := store.CreateJob(context.Background(), "key1", "job1", payload); err != nil {
		t.Fatalf("CreateJob error: %v", err)
	}
	if err := store.CreateJob(context.Background(), "key1", "job2", payload); !errors.Is(err, storeerr.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestStore_Unavailable(t *testing.T) {
	store, mr := newTestStore(t)
	mr.Close()

	payload := json.RawMessage(`{"a":1}`)
	err := store.CreateJob(context.Background(), "key1", "job1", payload)
	if !errors.Is(err, storeerr.ErrStoreUnavailable) {
		t.Fatalf("expected ErrStoreUnavailable, got %v", err)
	}
}
