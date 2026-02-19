package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	yaml "github.com/goccy/go-yaml"

	"mq-redis/internal/kafka"
	"mq-redis/internal/postgres"
	"mq-redis/internal/saga"
)

type Config struct {
	API             APIConfig       `yaml:"api"`
	Worker          WorkerConfig    `yaml:"worker"`
	RetryDispatcher RetryConfig     `yaml:"retry_dispatcher"`
	Redis           RedisConfig     `yaml:"redis"`
	Kafka           kafka.Config    `yaml:"kafka"`
	Postgres        postgres.Config `yaml:"postgres"`
	Saga            saga.Config     `yaml:"saga"`
}

type APIConfig struct {
	Addr string `yaml:"addr"`
}

type WorkerConfig struct {
	GroupID     string `yaml:"group_id"`
	Concurrency int    `yaml:"concurrency"`
}

type RetryConfig struct {
	PollInterval time.Duration `yaml:"poll_interval"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	return Parse(data)
}

func Parse(data []byte) (Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	cfg.applyDefaults()
	return cfg, nil
}

func (c *Config) applyDefaults() {
	if strings.TrimSpace(c.API.Addr) == "" {
		c.API.Addr = ":8080"
	}
	if strings.TrimSpace(c.Worker.GroupID) == "" {
		c.Worker.GroupID = "mq-worker"
	}
	if c.Worker.Concurrency <= 0 {
		c.Worker.Concurrency = 1
	}
	if c.RetryDispatcher.PollInterval <= 0 {
		c.RetryDispatcher.PollInterval = 1 * time.Second
	}
}

func (c Config) ValidateForAPI() error {
	if strings.TrimSpace(c.API.Addr) == "" {
		return fmt.Errorf("api.addr is required")
	}
	if err := validateRedis(c.Redis); err != nil {
		return err
	}
	if err := c.Kafka.ValidateJobs(); err != nil {
		return err
	}
	return nil
}

func (c Config) ValidateForWorker() error {
	if strings.TrimSpace(c.Worker.GroupID) == "" {
		return fmt.Errorf("worker.group_id is required")
	}
	if err := validateRedis(c.Redis); err != nil {
		return err
	}
	if err := c.Kafka.ValidateJobs(); err != nil {
		return err
	}
	return nil
}

func (c Config) ValidateForRetryDispatcher() error {
	if c.RetryDispatcher.PollInterval <= 0 {
		return fmt.Errorf("retry_dispatcher.poll_interval is required")
	}
	if err := validateRedis(c.Redis); err != nil {
		return err
	}
	if err := c.Kafka.ValidateJobs(); err != nil {
		return err
	}
	return nil
}

func validateRedis(cfg RedisConfig) error {
	if strings.TrimSpace(cfg.Addr) == "" {
		return fmt.Errorf("redis.addr is required")
	}
	return nil
}
