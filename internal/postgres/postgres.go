package postgres

import (
	"fmt"
	"strings"
)

type Config struct {
	DSN string `yaml:"dsn"`
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.DSN) == "" {
		return fmt.Errorf("postgres.dsn is required")
	}
	return nil
}
