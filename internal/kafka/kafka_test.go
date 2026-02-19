package kafka

import (
	"context"
	"errors"
	"io"
	"testing"

	segkafka "github.com/segmentio/kafka-go"
)

type fakeCloser struct {
	closed bool
}

func (c *fakeCloser) Close() error {
	c.closed = true
	return nil
}

type fakeWriter struct {
	msgs []segkafka.Message
	Err  error
}

func (w *fakeWriter) WriteMessages(ctx context.Context, msgs ...segkafka.Message) error {
	w.msgs = append(w.msgs, msgs...)
	return w.Err
}

func (w *fakeWriter) Close() error {
	return nil
}

func TestValidateJobs(t *testing.T) {
	cfg := Config{Brokers: []string{"b1"}, JobsTopic: "jobs"}
	if err := cfg.ValidateJobs(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestValidateJobsMissing(t *testing.T) {
	cfg := Config{}
	if err := cfg.ValidateJobs(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckConnectivityNoBrokers(t *testing.T) {
	if err := checkConnectivity(context.Background(), nil, func(ctx context.Context, network, address string) (io.Closer, error) {
		return nil, nil
	}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckConnectivityDialError(t *testing.T) {
	expected := errors.New("dial failed")
	err := checkConnectivity(context.Background(), []string{"b1"}, func(ctx context.Context, network, address string) (io.Closer, error) {
		return nil, expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected dial error")
	}
}

func TestCheckConnectivitySuccess(t *testing.T) {
	closer := &fakeCloser{}
	err := checkConnectivity(context.Background(), []string{"b1"}, func(ctx context.Context, network, address string) (io.Closer, error) {
		return closer, nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !closer.closed {
		t.Fatalf("expected close")
	}
}

func TestKafkaGoProducerPublish(t *testing.T) {
	w := &fakeWriter{}
	p := newKafkaGoProducerWithWriter(w)
	err := p.Publish(context.Background(), "topic", Message{Key: "k1", Value: []byte("v1")})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(w.msgs) != 1 {
		t.Fatalf("expected 1 message")
	}
	msg := w.msgs[0]
	if msg.Topic != "topic" {
		t.Fatalf("topic = %q", msg.Topic)
	}
	if string(msg.Key) != "k1" {
		t.Fatalf("key = %q", msg.Key)
	}
	if string(msg.Value) != "v1" {
		t.Fatalf("value = %q", msg.Value)
	}
}
