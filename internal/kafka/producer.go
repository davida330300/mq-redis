package kafka

import (
	"context"
	"fmt"

	segkafka "github.com/segmentio/kafka-go"
)

type writer interface {
	WriteMessages(ctx context.Context, msgs ...segkafka.Message) error
	Close() error
}

type KafkaGoProducer struct {
	writer writer
}

func NewKafkaGoProducer(cfg Config) (*KafkaGoProducer, error) {
	if err := cfg.ValidateJobs(); err != nil {
		return nil, err
	}
	w := &segkafka.Writer{
		Addr:     segkafka.TCP(cfg.Brokers...),
		Balancer: &segkafka.LeastBytes{},
	}
	return &KafkaGoProducer{writer: w}, nil
}

func newKafkaGoProducerWithWriter(w writer) *KafkaGoProducer {
	return &KafkaGoProducer{writer: w}
}

func (p *KafkaGoProducer) Publish(ctx context.Context, topic string, msg Message) error {
	if p == nil || p.writer == nil {
		return fmt.Errorf("kafka producer not configured")
	}
	return p.writer.WriteMessages(ctx, segkafka.Message{
		Topic: topic,
		Key:   []byte(msg.Key),
		Value: msg.Value,
	})
}

func (p *KafkaGoProducer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
