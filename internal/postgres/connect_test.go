package postgres

import (
	"context"
	"errors"
	"testing"
)

func TestCheckConnectivityMissingDSN(t *testing.T) {
	if err := checkConnectivity(context.Background(), "", func(ctx context.Context, dsn string) error {
		return nil
	}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckConnectivityPingError(t *testing.T) {
	expected := errors.New("ping failed")
	err := checkConnectivity(context.Background(), "dsn", func(ctx context.Context, dsn string) error {
		return expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected error")
	}
}

func TestCheckConnectivitySuccess(t *testing.T) {
	err := checkConnectivity(context.Background(), "dsn", func(ctx context.Context, dsn string) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
