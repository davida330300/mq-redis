package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"mq-redis/internal/config"
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
	if err := cfg.ValidateForRetryDispatcher(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	log.Printf("retry-dispatcher starting poll_interval=%s", cfg.RetryDispatcher.PollInterval)
	log.Printf("retry-dispatcher using redis=%s kafka_brokers=%v", cfg.Redis.Addr, cfg.Kafka.Brokers)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Printf("retry-dispatcher shutting down")
}
