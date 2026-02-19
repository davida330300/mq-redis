package kafka

import (
	"context"
	"fmt"
	"io"
	"time"

	segkafka "github.com/segmentio/kafka-go"
)

type dialFunc func(ctx context.Context, network, address string) (io.Closer, error)

func CheckConnectivity(ctx context.Context, brokers []string) error {
	return checkConnectivity(ctx, brokers, defaultDialer())
}

func checkConnectivity(ctx context.Context, brokers []string, dial dialFunc) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers configured")
	}
	conn, err := dial(ctx, "tcp", brokers[0])
	if err != nil {
		return err
	}
	return conn.Close()
}

func defaultDialer() dialFunc {
	dialer := &segkafka.Dialer{Timeout: 2 * time.Second}
	return func(ctx context.Context, network, address string) (io.Closer, error) {
		return dialer.DialContext(ctx, network, address)
	}
}
