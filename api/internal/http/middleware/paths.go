package middleware

import (
	"net/http"
	"strings"
)

func isProtectedPath(r *http.Request) bool {
	if r == nil || r.URL == nil {
		return false
	}
	path := r.URL.Path
	if path == "/v1/license" || path == "/api/v1/license" {
		return false
	}
	return path == "/v1" ||
		strings.HasPrefix(path, "/v1/") ||
		path == "/api/v1" ||
		strings.HasPrefix(path, "/api/v1/")
}

func isAdminPath(r *http.Request) bool {
	if r == nil || r.URL == nil {
		return false
	}
	path := r.URL.Path
	return path == "/v1/admin" ||
		strings.HasPrefix(path, "/v1/admin/") ||
		path == "/api/v1/admin" ||
		strings.HasPrefix(path, "/api/v1/admin/")
}
