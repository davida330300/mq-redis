package retry

import (
	"errors"
	"math/rand"
	"time"
)

type Config struct {
	Base   time.Duration
	Max    time.Duration
	Jitter float64
}

func DefaultConfig() Config {
	return Config{
		Base:   1 * time.Second,
		Max:    60 * time.Second,
		Jitter: 0.2,
	}
}

func (c Config) Validate() error {
	if c.Base <= 0 {
		return errors.New("base must be positive")
	}
	if c.Max <= 0 {
		return errors.New("max must be positive")
	}
	if c.Max < c.Base {
		return errors.New("max must be >= base")
	}
	if c.Jitter < 0 || c.Jitter >= 1 {
		return errors.New("jitter must be in [0,1)")
	}
	return nil
}

func NextDelay(cfg Config, attempt int64, rng *rand.Rand) (time.Duration, error) {
	if err := cfg.Validate(); err != nil {
		return 0, err
	}
	if attempt < 1 {
		return 0, errors.New("attempt must be >= 1")
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	delay := cfg.Base
	for i := int64(1); i < attempt; i++ {
		if delay >= cfg.Max/2 {
			delay = cfg.Max
			break
		}
		delay *= 2
	}
	if delay > cfg.Max {
		delay = cfg.Max
	}

	if cfg.Jitter > 0 {
		jitterRange := cfg.Jitter * 2
		delta := (rng.Float64() * jitterRange) - cfg.Jitter
		jittered := float64(delay) * (1 + delta)
		if jittered < float64(time.Millisecond) {
			jittered = float64(time.Millisecond)
		}
		delay = time.Duration(jittered)
	}
	return delay, nil
}

func NextScore(now time.Time, delay time.Duration) float64 {
	return float64(now.Add(delay).UnixMilli())
}
