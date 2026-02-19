package saga

import (
	"fmt"
)

type Config struct {
	Enabled bool `yaml:"enabled"`
}

type Step struct {
	Name string
}

type Definition struct {
	Steps []Step
}

func (d Definition) Validate() error {
	if len(d.Steps) == 0 {
		return fmt.Errorf("saga requires at least one step")
	}
	seen := make(map[string]struct{}, len(d.Steps))
	for _, step := range d.Steps {
		if step.Name == "" {
			return fmt.Errorf("saga step name is required")
		}
		if _, exists := seen[step.Name]; exists {
			return fmt.Errorf("duplicate saga step: %s", step.Name)
		}
		seen[step.Name] = struct{}{}
	}
	return nil
}
