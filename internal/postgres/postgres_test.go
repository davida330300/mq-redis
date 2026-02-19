package postgres

import "testing"

func TestValidate(t *testing.T) {
	cfg := Config{DSN: "postgres://user:pass@localhost:5432/db"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestValidateMissing(t *testing.T) {
	cfg := Config{}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error")
	}
}
