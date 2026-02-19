package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"mq-redis/internal/kafka"
	"mq-redis/internal/rediskeys"
	"mq-redis/internal/state"
)

type Processor interface {
	Process(ctx context.Context, jobID string, payload json.RawMessage) error
}

type NoopProcessor struct{}

func (p *NoopProcessor) Process(ctx context.Context, jobID string, payload json.RawMessage) error {
	return nil
}

type Worker struct {
	consumer    kafka.Consumer
	dlqProducer kafka.Producer
	retryDelay  time.Duration
	now         func() time.Time
	dlqTopic    string
	redis       *redis.Client
	processor   Processor
}

func New(consumer kafka.Consumer, redisClient *redis.Client, processor Processor, dlqProducer kafka.Producer, dlqTopic string) (*Worker, error) {
	if consumer == nil {
		return nil, errors.New("consumer is required")
	}
	if redisClient == nil {
		return nil, errors.New("redis client is required")
	}
	if processor == nil {
		return nil, errors.New("processor is required")
	}
	return &Worker{
		consumer:    consumer,
		dlqProducer: dlqProducer,
		dlqTopic:    dlqTopic,
		retryDelay:  1 * time.Second,
		now:         time.Now,
		redis:       redisClient,
		processor:   processor,
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		msg, err := w.consumer.Poll(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("worker poll error: %v", err)
			continue
		}
		if err := w.Handle(ctx, msg); err != nil {
			log.Printf("worker handle error: %v", err)
		}
	}
}

func (w *Worker) Handle(ctx context.Context, msg kafka.Message) error {
	jobID := msg.Key
	if jobID == "" {
		return errors.New("missing job id")
	}

	w.setStatus(ctx, jobID, state.Processing, rediskeys.JobStatusTTL)

	if err := w.processor.Process(ctx, jobID, msg.Value); err == nil {
		w.setStatus(ctx, jobID, state.Done, rediskeys.JobStatusTTL)
		return nil
	}

	attempt, err := w.bumpAttempt(ctx, jobID)
	if err != nil {
		log.Printf("attempt increment failed: %v", err)
	}

	if attempt <= 1 {
		w.setStatus(ctx, jobID, state.Retrying, rediskeys.JobStatusTTL)
		w.scheduleRetry(ctx, jobID)
		return errors.New("job failed; scheduled retry")
	}

	w.setStatus(ctx, jobID, state.DLQ, rediskeys.DLQTTL)
	if w.dlqProducer != nil && w.dlqTopic != "" {
		if err := w.dlqProducer.Publish(ctx, w.dlqTopic, kafka.Message{Key: jobID, Value: msg.Value}); err != nil {
			log.Printf("dlq publish failed: %v", err)
		}
	}
	return errors.New("job failed; sent to dlq")
}

func (w *Worker) bumpAttempt(ctx context.Context, jobID string) (int64, error) {
	key := rediskeys.AttemptKey(jobID)
	attempt, err := w.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if attempt == 1 {
		if err := w.redis.Expire(ctx, key, rediskeys.JobDataTTL).Err(); err != nil {
			log.Printf("attempt ttl set failed: %v", err)
		}
	}
	return attempt, nil
}

func (w *Worker) setStatus(ctx context.Context, jobID string, status state.State, ttl time.Duration) {
	if err := w.redis.Set(ctx, rediskeys.JobKey(jobID), string(status), ttl).Err(); err != nil {
		log.Printf("status update failed: %v", err)
	}
}

func (w *Worker) scheduleRetry(ctx context.Context, jobID string) {
	score := float64(w.now().Add(w.retryDelay).UnixMilli())
	if err := w.redis.ZAdd(ctx, rediskeys.RetryJobsKey, redis.Z{Score: score, Member: jobID}).Err(); err != nil {
		log.Printf("retry schedule failed: %v", err)
	}
}
