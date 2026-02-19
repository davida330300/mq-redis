package retry

import (
	"math/rand"
	"testing"
	"time"
)

func TestNextDelayBase(t *testing.T) {
	cfg := Config{Base: 1 * time.Second, Max: 10 * time.Second, Jitter: 0}
	delay, err := NextDelay(cfg, 1, rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if delay != 1*time.Second {
		t.Fatalf("delay = %v", delay)
	}
}

func TestNextDelayExponent(t *testing.T) {
	cfg := Config{Base: 1 * time.Second, Max: 10 * time.Second, Jitter: 0}
	delay, err := NextDelay(cfg, 3, rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if delay != 4*time.Second {
		t.Fatalf("delay = %v", delay)
	}
}

func TestNextDelayCap(t *testing.T) {
	cfg := Config{Base: 2 * time.Second, Max: 5 * time.Second, Jitter: 0}
	delay, err := NextDelay(cfg, 4, rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if delay != 5*time.Second {
		t.Fatalf("delay = %v", delay)
	}
}

func TestNextDelayJitterRange(t *testing.T) {
	cfg := Config{Base: 10 * time.Second, Max: 10 * time.Second, Jitter: 0.2}
	rng := rand.New(rand.NewSource(42))
	delay, err := NextDelay(cfg, 1, rng)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	min := 8 * time.Second
	max := 12 * time.Second
	if delay < min || delay > max {
		t.Fatalf("delay out of range: %v", delay)
	}
}

func TestNextDelayInvalidAttempt(t *testing.T) {
	cfg := Config{Base: 1 * time.Second, Max: 10 * time.Second, Jitter: 0}
	if _, err := NextDelay(cfg, 0, rand.New(rand.NewSource(1))); err == nil {
		t.Fatalf("expected error")
	}
}

func TestNextScore(t *testing.T) {
	when := time.Unix(0, 0)
	score := NextScore(when, 1500*time.Millisecond)
	if score != 1500 {
		t.Fatalf("score = %v", score)
	}
}
