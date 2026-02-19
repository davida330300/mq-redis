package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pingFunc func(ctx context.Context, dsn string) error

func CheckConnectivity(ctx context.Context, dsn string) error {
	return checkConnectivity(ctx, dsn, defaultPing)
}

func checkConnectivity(ctx context.Context, dsn string, ping pingFunc) error {
	if dsn == "" {
		return fmt.Errorf("postgres dsn is empty")
	}
	return ping(ctx, dsn)
}

func defaultPing(ctx context.Context, dsn string) error {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()
	return pool.Ping(ctx)
}
