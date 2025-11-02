package bandit

import (
	"strings"
)

// ExperimentConfig controls optional exploration experiments.
type ExperimentConfig struct {
	Enabled        bool
	HoldoutPercent float64
	Label          string
	Surfaces       map[string]struct{}
}

// Normalized returns a sanitized copy of the config.
func (c ExperimentConfig) Normalized() ExperimentConfig {
	if c.HoldoutPercent < 0 {
		c.HoldoutPercent = 0
	}
	if c.HoldoutPercent > 1 {
		c.HoldoutPercent = 1
	}
	if c.Label == "" {
		c.Label = "bandit_experiment"
	}
	if len(c.Surfaces) > 0 {
		out := make(map[string]struct{}, len(c.Surfaces))
		for k := range c.Surfaces {
			key := strings.ToLower(strings.TrimSpace(k))
			if key == "" {
				continue
			}
			out[key] = struct{}{}
		}
		c.Surfaces = out
	}
	c.Enabled = c.Enabled && c.HoldoutPercent > 0
	return c
}

// Applies reports whether the experiment applies to the given surface.
func (c ExperimentConfig) Applies(surface string) bool {
	if !c.Enabled {
		return false
	}
	if len(c.Surfaces) == 0 {
		return true
	}
	key := strings.ToLower(strings.TrimSpace(surface))
	if key == "" {
		key = "default"
	}
	_, ok := c.Surfaces[key]
	return ok
}
