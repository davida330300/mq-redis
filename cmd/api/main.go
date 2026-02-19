package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"mq-redis/internal/api"
	"mq-redis/internal/config"
	producerkafka "mq-redis/internal/producer/kafka"
	redisstore "mq-redis/internal/store/redis"
)

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

	store := redisstore.New(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("redis close error: %v", err)
		}
	}()

	producer, err := producerkafka.New(cfg.Kafka, nil)
	if err != nil {
		log.Fatalf("kafka producer: %v", err)
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
