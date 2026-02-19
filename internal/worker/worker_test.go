package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"mq-redis/internal/kafka"
	"mq-redis/internal/rediskeys"
)

type fakeProcessor struct {
	err error
}

func (p *fakeProcessor) Process(ctx context.Context, jobID string, payload json.RawMessage) error {
	return p.err
}

type fakeDLQProducer struct {
	msgs []kafka.Message
}

func (p *fakeDLQProducer) Publish(ctx context.Context, topic string, msg kafka.Message) error {
	p.msgs = append(p.msgs, msg)
	return nil
}

func TestHandleSuccess(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	worker, err := New(&fakeConsumer{}, client, &fakeProcessor{}, nil, "")
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}

	msg := kafka.Message{Key: "job1", Value: []byte(`{"a":1}`)}
	if err := worker.Handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}

	status, err := client.Get(context.Background(), rediskeys.JobKey("job1")).Result()
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status != "done" {
		t.Fatalf("status = %q", status)
	}
	if client.Exists(context.Background(), rediskeys.AttemptKey("job1")).Val() != 0 {
		t.Fatalf("unexpected attempt key")
	}
}

func TestHandleRetryThenDLQ(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	dlq := &fakeDLQProducer{}
	worker, err := New(&fakeConsumer{}, client, &fakeProcessor{err: errors.New("boom")}, dlq, "jobs.dlq")
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	worker.now = func() time.Time { return time.Unix(0, 0) }
	worker.retryDelay = 2 * time.Second

	msg := kafka.Message{Key: "job1", Value: []byte(`{"a":1}`)}
	_ = worker.Handle(context.Background(), msg)

	status, _ := client.Get(context.Background(), rediskeys.JobKey("job1")).Result()
	if status != "retrying" {
		t.Fatalf("status = %q", status)
	}

	attempt, _ := client.Get(context.Background(), rediskeys.AttemptKey("job1")).Result()
	if attempt != "1" {
		t.Fatalf("attempt = %q", attempt)
	}

	members, err := mr.ZMembers(rediskeys.RetryJobsKey)
	if err != nil {
		t.Fatalf("retry members: %v", err)
	}
	if len(members) != 1 || members[0] != "job1" {
		t.Fatalf("retry members = %v", members)
	}

	_ = worker.Handle(context.Background(), msg)
	status, _ = client.Get(context.Background(), rediskeys.JobKey("job1")).Result()
	if status != "dlq" {
		t.Fatalf("status = %q", status)
	}
	if len(dlq.msgs) != 1 {
		t.Fatalf("expected dlq publish")
	}
}

type fakeConsumer struct{}

func (c *fakeConsumer) Poll(ctx context.Context) (kafka.Message, error) {
	return kafka.Message{}, errors.New("nope")
}

func (c *fakeConsumer) Commit(ctx context.Context, msg kafka.Message) error {
	return nil
}

func (c *fakeConsumer) Close() error {
	return nil
}
