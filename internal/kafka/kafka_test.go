package kafka

import "testing"

func TestValidateJobs(t *testing.T) {
	cfg := Config{Brokers: []string{"b1"}, JobsTopic: "jobs"}
	if err := cfg.ValidateJobs(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestValidateJobsMissing(t *testing.T) {
	cfg := Config{}
	if err := cfg.ValidateJobs(); err == nil {
		t.Fatalf("expected error")
	}
}
