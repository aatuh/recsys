package problem

import (
	"net/http"
	"strings"

	"github.com/aatuh/api-toolkit/v2/httpx"
	"github.com/go-chi/chi/v5/middleware"
)

const typeBase = "https://errors.recsys.example/problem/"

// Write emits a Problem Details response with extended fields.
func Write(w http.ResponseWriter, r *http.Request, status int, code, detail string) {
	if w == nil {
		return
	}
	if code == "" {
		code = "RECSYS_ERROR"
	}
	problemType := typeBase + strings.ToLower(strings.ReplaceAll(code, "_", "-"))
	p := httpx.Problem{
		Type:     problemType,
		Title:    http.StatusText(status),
		Detail:   detail,
		Instance: instancePath(r),
	}
	p.With("code", code)
	if reqID := RequestID(r); reqID != "" {
		p.With("request_id", reqID)
	}
	SetRequestIDHeader(w, r)
	httpx.WriteProblem(w, status, p)
}

// RequestID returns the request identifier from headers or context.
func RequestID(r *http.Request) string {
	if r == nil {
		return ""
	}
	if v := strings.TrimSpace(r.Header.Get("X-Request-Id")); v != "" {
		return v
	}
	if v := strings.TrimSpace(r.Header.Get("X-Request-ID")); v != "" {
		return v
	}
	if v := middleware.GetReqID(r.Context()); v != "" {
		return v
	}
	return ""
}

// SetRequestIDHeader ensures the response includes the request id header.
func SetRequestIDHeader(w http.ResponseWriter, r *http.Request) {
	if w == nil {
		return
	}
	if reqID := RequestID(r); reqID != "" {
		w.Header().Set("X-Request-Id", reqID)
	}
}

func instancePath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	return r.URL.Path
}
