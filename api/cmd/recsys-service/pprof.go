package main

import (
	"net/http"

	"github.com/aatuh/recsys-suite/api/internal/config"
)

func systemPprofHandler(cfg config.Config) http.Handler {
	if !cfg.Performance.PprofEnabled {
		return nil
	}
	return http.DefaultServeMux
}
