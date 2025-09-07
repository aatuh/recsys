package util

import (
	"fmt"
	"os"
)

func MustGetEnv(k string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	panic(fmt.Sprintf("mustgetenv: %s is not set", k))
}

func GetEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
