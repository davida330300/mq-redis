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
	"mq-redis/internal/worker"
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

	consumer, err := kafka.NewKafkaGoConsumer(cfg.Kafka, cfg.Worker.GroupID)
	if err != nil {
		log.Fatalf("kafka consumer init failed: %v", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("kafka consumer close error: %v", err)
		}
	}()

	dlqProducer, err := kafka.NewKafkaGoProducer(cfg.Kafka)
	if err != nil {
		log.Printf("kafka dlq producer init failed: %v", err)
	}
	if dlqProducer != nil {
		defer func() {
			if err := dlqProducer.Close(); err != nil {
				log.Printf("kafka dlq producer close error: %v", err)
			}
		}()
	}

	runner, err := worker.New(consumer, redisClient, &worker.NoopProcessor{}, dlqProducer, cfg.Kafka.DLQTopic)
	if err != nil {
		log.Fatalf("worker init failed: %v", err)
	}

	runCtx, cancelRun := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Run(runCtx)
	}()

	log.Printf("worker starting group=%s concurrency=%d", cfg.Worker.GroupID, cfg.Worker.Concurrency)
	log.Printf("worker using redis=%s kafka_brokers=%v", cfg.Redis.Addr, cfg.Kafka.Brokers)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	cancelRun()
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			log.Printf("worker stopped with error: %v", err)
		}
	case <-time.After(3 * time.Second):
		log.Printf("worker shutdown timed out")
	}
	log.Printf("worker shutting down")
}
