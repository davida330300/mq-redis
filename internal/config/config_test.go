package config

import "testing"

func TestParseDefaults(t *testing.T) {
	cfg, err := Parse([]byte(`redis:
  addr: "localhost:6379"
  db: 0
kafka:
  brokers: ["localhost:9092"]
  jobs_topic: "jobs"
`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.API.Addr != ":8080" {
		t.Fatalf("api.addr default = %q", cfg.API.Addr)
	}
	if cfg.Worker.GroupID != "mq-worker" {
		t.Fatalf("worker.group_id default = %q", cfg.Worker.GroupID)
	}
	if cfg.Worker.Concurrency != 1 {
		t.Fatalf("worker.concurrency default = %d", cfg.Worker.Concurrency)
	}
	if err := cfg.ValidateForAPI(); err != nil {
		t.Fatalf("validate for api: %v", err)
	}
}

func TestValidateForAPIRequiresRedis(t *testing.T) {
	cfg, err := Parse([]byte(`kafka:
  brokers: ["localhost:9092"]
  jobs_topic: "jobs"
`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := cfg.ValidateForAPI(); err == nil {
		t.Fatalf("expected error")
	}
}
