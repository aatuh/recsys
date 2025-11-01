package handlers

import (
	"net/http"

	"github.com/google/uuid"
)

func orgIDFromHeader(r *http.Request, fallback uuid.UUID) uuid.UUID {
	if r == nil {
		return fallback
	}
	org := r.Header.Get("X-Org-ID")
	if id, err := uuid.Parse(org); err == nil {
		return id
	}
	return fallback
}
