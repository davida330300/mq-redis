package kafka

import (
	"context"
	"fmt"
	"strings"
)

type Config struct {
	Brokers   []string `yaml:"brokers"`
	JobsTopic string   `yaml:"jobs_topic"`
	DLQTopic  string   `yaml:"dlq_topic"`
	ClientID  string   `yaml:"client_id"`
}

func (c Config) ValidateJobs() error {
	if len(c.Brokers) == 0 {
		return fmt.Errorf("kafka.brokers is required")
	}
	if strings.TrimSpace(c.JobsTopic) == "" {
		return fmt.Errorf("kafka.jobs_topic is required")
	}
	return nil
}

type Message struct {
	Key   string
	Value []byte
}

type Producer interface {
	Publish(ctx context.Context, topic string, msg Message) error
}

type Consumer interface {
	Poll(ctx context.Context) (Message, error)
	Commit(ctx context.Context, msg Message) error
	Close() error
}

type NoopProducer struct{}

func (p *NoopProducer) Publish(ctx context.Context, topic string, msg Message) error {
	return nil
}
