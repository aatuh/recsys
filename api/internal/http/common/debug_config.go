package common

import "strings"

type DebugConfig struct {
	Environment string
	AppDebug    bool
}

// NewDebugConfig normalises environment metadata.
func NewDebugConfig(environment string, appDebug bool) DebugConfig {
	return DebugConfig{
		Environment: strings.ToLower(strings.TrimSpace(environment)),
		AppDebug:    appDebug,
	}
}

func (c DebugConfig) IsDebug() bool {
	return c.Environment == "dev" ||
		c.Environment == "development" ||
		c.Environment == "local" ||
		c.AppDebug
}
