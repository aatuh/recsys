package config

import "os"

// Source defines how configuration values are looked up.
type Source interface {
	Lookup(key string) (string, bool)
}

// EnvSource reads configuration values from process environment variables.
type EnvSource struct{}

// Lookup implements Source using os.LookupEnv.
func (EnvSource) Lookup(key string) (string, bool) {
	return os.LookupEnv(key)
}
