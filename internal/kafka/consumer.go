package kafka

import (
	"context"
	"fmt"

	segkafka "github.com/segmentio/kafka-go"
)

type KafkaGoConsumer struct {
	reader *segkafka.Reader
}

func NewKafkaGoConsumer(cfg Config, groupID string) (*KafkaGoConsumer, error) {
	if err := cfg.ValidateJobs(); err != nil {
		return nil, err
	}
	if groupID == "" {
		return nil, fmt.Errorf("groupID is required")
	}
	reader := segkafka.NewReader(segkafka.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.JobsTopic,
		GroupID: groupID,
	})
	return &KafkaGoConsumer{reader: reader}, nil
}

func (c *KafkaGoConsumer) Poll(ctx context.Context) (Message, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return Message{}, err
	}
	return Message{Key: string(msg.Key), Value: msg.Value}, nil
}

func (c *KafkaGoConsumer) Commit(ctx context.Context, msg Message) error {
	return nil
}

func (c *KafkaGoConsumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}
