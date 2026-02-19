package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"mq-redis/internal/config"
	"mq-redis/internal/kafka"
	"mq-redis/internal/postgres"
)

const connectTimeout = 2 * time.Second

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config/config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := cfg.ValidateForWorker(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("redis close error: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("redis ping failed: %v", err)
	}
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), connectTimeout)
	if err := kafka.CheckConnectivity(ctx, cfg.Kafka.Brokers); err != nil {
		log.Printf("kafka connectivity check failed: %v", err)
	}
	cancel()

	if cfg.Postgres.DSN == "" {
		log.Printf("postgres dsn missing; skipping connectivity check")
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), connectTimeout)
		if err := postgres.CheckConnectivity(ctx, cfg.Postgres.DSN); err != nil {
			log.Printf("postgres connectivity check failed: %v", err)
		}
		cancel()
	}

	log.Printf("worker starting group=%s concurrency=%d", cfg.Worker.GroupID, cfg.Worker.Concurrency)
	log.Printf("worker using redis=%s kafka_brokers=%v", cfg.Redis.Addr, cfg.Kafka.Brokers)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Printf("worker shutting down")
}
