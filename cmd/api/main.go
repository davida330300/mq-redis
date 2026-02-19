package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"mq-redis/internal/api"
	"mq-redis/internal/config"
	"mq-redis/internal/kafka"
	"mq-redis/internal/postgres"
	producerkafka "mq-redis/internal/producer/kafka"
	redisstore "mq-redis/internal/store/redis"
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
	if err := cfg.ValidateForAPI(); err != nil {
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

	store := redisstore.NewWithClient(redisClient)
	producer, err := producerkafka.New(cfg.Kafka, nil)
	if err != nil {
		log.Printf("kafka producer init failed: %v", err)
	}
	if producer != nil {
		defer func() {
			if err := producer.Close(); err != nil {
				log.Printf("kafka producer close error: %v", err)
			}
		}()
	}

	r := api.NewRouter(store, producer)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, api.JobResponse{Status: "ok"})
	})

	server := &http.Server{
		Addr:    cfg.API.Addr,
		Handler: r,
	}

	log.Printf("api listening on %s", cfg.API.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
