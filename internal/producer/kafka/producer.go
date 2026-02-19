package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"mq-redis/internal/kafka"
)

type Producer struct {
	topic    string
	producer kafka.Producer
}

func New(cfg kafka.Config, producer kafka.Producer) (*Producer, error) {
	if err := cfg.ValidateJobs(); err != nil {
		return nil, err
	}
	if producer == nil {
		real, err := kafka.NewKafkaGoProducer(cfg)
		if err != nil {
			return nil, err
		}
		producer = real
	}
	return &Producer{topic: cfg.JobsTopic, producer: producer}, nil
}

func (p *Producer) Publish(ctx context.Context, jobID string, payload json.RawMessage) error {
	if p == nil || p.producer == nil {
		return fmt.Errorf("kafka producer not configured")
	}
	msg := kafka.Message{Key: jobID, Value: payload}
	return p.producer.Publish(ctx, p.topic, msg)
}

func (p *Producer) Close() error {
	if p == nil || p.producer == nil {
		return nil
	}
	if closer, ok := p.producer.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}
