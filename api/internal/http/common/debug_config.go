package common

import (
	"recsys/shared/util"
	"strings"
)

type DebugConfig struct {
	Environment string
	AppDebug    bool
}

func LoadDebugConfig() DebugConfig {
	env := strings.ToLower(util.MustGetEnv("ENV"))
	appDebug := util.MustGetEnv("APP_DEBUG") == "true"

	return DebugConfig{
		Environment: env,
		AppDebug:    appDebug,
	}
}

func (c DebugConfig) IsDebug() bool {
	return c.Environment == "dev" ||
		c.Environment == "development" ||
		c.Environment == "local" ||
		c.AppDebug
}
