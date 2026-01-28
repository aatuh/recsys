package handlers

import (
	"net/http"
	"strings"

	"recsys/internal/validation"

	"github.com/aatuh/api-toolkit/httpx"
	"github.com/go-chi/chi/v5/middleware"
)

const errorTypeBase = "https://errors.recsys.example/problem/"

func writeValidationError(w http.ResponseWriter, r *http.Request, err error) {
	if verr, ok := err.(validation.Error); ok {
		writeProblem(w, r, verr.Status, verr.Code, verr.Message)
		return
	}
	writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, detail string) {
	problemType := errorTypeBase + strings.ToLower(strings.ReplaceAll(code, "_", "-"))
	reqID := requestIDFromRequest(r)
	p := httpx.Problem{
		Type:     problemType,
		Title:    http.StatusText(status),
		Detail:   detail,
		Instance: instancePath(r),
	}
	p.With("code", code)
	if reqID != "" {
		p.With("request_id", reqID)
	}
	setRequestIDHeader(w, r)
	httpx.WriteProblem(w, status, p)
}

func requestIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if v := strings.TrimSpace(r.Header.Get("X-Request-Id")); v != "" {
		return v
	}
	if v := middleware.GetReqID(r.Context()); v != "" {
		return v
	}
	return ""
}

func instancePath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	return r.URL.Path
}

func setRequestIDHeader(w http.ResponseWriter, r *http.Request) {
	if w == nil {
		return
	}
	if reqID := requestIDFromRequest(r); reqID != "" {
		w.Header().Set("X-Request-Id", reqID)
	}
}
