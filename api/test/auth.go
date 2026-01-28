package test

import "net/http"

// applyAuthHeaders attaches auth headers for integration tests.
func applyAuthHeaders(req *http.Request, cfg Config) {
	if req == nil {
		return
	}
	if cfg.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)
	}
	if cfg.DevAuth.Enabled {
		if cfg.DevAuth.UserIDHeader != "" {
			req.Header.Set(cfg.DevAuth.UserIDHeader, cfg.DevAuth.UserIDValue)
		}
		if cfg.DevAuth.TenantHeader != "" {
			req.Header.Set(cfg.DevAuth.TenantHeader, cfg.DevAuth.TenantIDValue)
		}
	}
}
