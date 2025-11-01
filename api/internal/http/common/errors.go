package common

import (
	"encoding/json"
	"errors"
	"net/http"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type APIError struct {
	Code          string         `json:"error"`
	Message       string         `json:"message,omitempty"`
	Details       map[string]any `json:"details,omitempty"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	Timestamp     time.Time      `json:"timestamp,omitempty"`
	httpStatus    int            `json:"-"`
	debugMsg      string         `json:"-"`
	stackTrace    string         `json:"-"`
}

// ErrorContext provides additional context for error logging
type ErrorContext struct {
	UserID    string
	OrgID     string
	RequestID string
	Path      string
	Method    string
	UserAgent string
	IP        string
}

func NewAPIError(code, msg string, status int) APIError {
	return APIError{
		Code:       code,
		Message:    msg,
		httpStatus: status,
		Timestamp:  time.Now().UTC(),
	}
}

// NewAPIErrorWithContext creates an API error with additional context
func NewAPIErrorWithContext(code, msg string, status int, ctx ErrorContext) APIError {
	ae := NewAPIError(code, msg, status)
	ae.CorrelationID = ctx.RequestID
	return ae
}

// WithStackTrace adds stack trace to error for debugging
func (ae APIError) WithStackTrace() APIError {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			ae.stackTrace = string(buf[:n])
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	return ae
}

// WithDebugMessage adds debug message to error
func (ae APIError) WithDebugMessage(msg string) APIError {
	ae.debugMsg = msg
	return ae
}

func BadRequest(w http.ResponseWriter, r *http.Request, code, msg string, details map[string]any) {
	ae := NewAPIError(code, msg, http.StatusBadRequest)
	ae.Details = details
	writeJSON(w, r, ae)
}
func NotFound(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, NewAPIError("not_found", "Not found", http.StatusNotFound))
}
func Unauthorized(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, NewAPIError("unauthorized", "Unauthorized", http.StatusUnauthorized))
}
func Forbidden(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, NewAPIError("forbidden", "Forbidden", http.StatusForbidden))
}
func ServiceUnavailable(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, NewAPIError("service_unavailable", "Service unavailable", http.StatusServiceUnavailable))
}
func TooManyRequests(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, NewAPIError("rate_limited", "Too many requests", http.StatusTooManyRequests))
}
func Unprocessable(w http.ResponseWriter, r *http.Request, code, msg string, details map[string]any) {
	ae := NewAPIError(code, msg, http.StatusUnprocessableEntity)
	ae.Details = details
	writeJSON(w, r, ae)
}

func HttpError(w http.ResponseWriter, r *http.Request, err error, fallback int) {
	ae := mapError(err)
	if ae.httpStatus == 0 {
		ae.httpStatus = fallback
	}

	// Add request context
	ae.CorrelationID = middleware.GetReqID(r.Context())

	// For 5xx errors, add stack trace in debug mode
	if ae.httpStatus >= 500 && isDebug() {
		ae = ae.WithStackTrace()
	}

	writeJSON(w, r, ae)
}

// HttpErrorWithLogger logs the error with structured logging
func HttpErrorWithLogger(w http.ResponseWriter, r *http.Request, err error, fallback int, logger *zap.Logger) {
	ae := mapError(err)
	if ae.httpStatus == 0 {
		ae.httpStatus = fallback
	}

	// Add request context
	ae.CorrelationID = middleware.GetReqID(r.Context())

	// Log the error with context
	logFields := []zap.Field{
		zap.String("error_code", ae.Code),
		zap.String("error_message", ae.Message),
		zap.Int("http_status", ae.httpStatus),
		zap.String("req_id", ae.CorrelationID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("user_agent", r.UserAgent()),
		zap.String("remote_addr", r.RemoteAddr),
	}

	if orgID := r.Header.Get("X-Org-ID"); orgID != "" {
		logFields = append(logFields, zap.String("org_id", orgID))
	}

	if ae.debugMsg != "" {
		logFields = append(logFields, zap.String("debug_msg", ae.debugMsg))
	}

	if ae.Details != nil {
		logFields = append(logFields, zap.Any("error_details", ae.Details))
	}

	// Use appropriate log level
	if ae.httpStatus >= 500 {
		logger.Error("http server error", append(logFields, zap.Error(err))...)
	} else {
		logger.Warn("http client error", append(logFields, zap.Error(err))...)
	}

	// For 5xx errors, add stack trace in debug mode
	if ae.httpStatus >= 500 && isDebug() {
		ae = ae.WithStackTrace()
	}

	writeJSON(w, r, ae)
}

func mapError(err error) APIError {
	if err == nil {
		return NewAPIError("internal", "Unexpected error", http.StatusInternalServerError)
	}

	// Malformed JSON
	var se *json.SyntaxError
	if errors.As(err, &se) {
		ae := NewAPIError("invalid_json", "Malformed JSON", http.StatusBadRequest)
		ae.debugMsg = err.Error()
		return ae
	}
	var ue *json.UnmarshalTypeError
	if errors.As(err, &ue) {
		ae := NewAPIError("invalid_json_type", "JSON has a value of the wrong type", http.StatusBadRequest)
		ae.Details = map[string]any{"field": ue.Field, "expected": ue.Type.String()}
		ae.debugMsg = err.Error()
		return ae
	}
	// Handle "unexpected EOF" and other JSON decode errors
	if err.Error() == "unexpected EOF" || err.Error() == "unexpected end of JSON input" {
		ae := NewAPIError("invalid_json", "Malformed JSON", http.StatusBadRequest)
		ae.debugMsg = err.Error()
		return ae
	}

	// Postgres constraint mapping (pgx)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return fromPgCode(pgErr.Code, err)
	}

	// Default
	ae := NewAPIError("internal", "Internal server error", http.StatusInternalServerError)
	ae.debugMsg = err.Error()
	return ae
}

// https://www.postgresql.org/docs/current/errcodes-appendix.html
func fromPgCode(code string, _ error) APIError {
	switch code {
	case "23505": // unique violation
		return NewAPIError("unique_violation", "Resource already exists", http.StatusConflict)
	case "23503": // foreign key violation
		return NewAPIError("foreign_key_violation", "Related resource does not exist", http.StatusUnprocessableEntity)
	case "23502": // not null violation
		return NewAPIError("not_null_violation", "Missing required field", http.StatusUnprocessableEntity)
	case "23514": // check violation
		return NewAPIError("check_violation", "Data violates a constraint", http.StatusUnprocessableEntity)
	case "22001": // string_data_right_truncation
		return NewAPIError("string_truncation", "String data right truncation", http.StatusUnprocessableEntity)
	default:
		return NewAPIError("constraint_violation", "Data violates a constraint", http.StatusUnprocessableEntity)
	}
}

// Global debug config - will be injected at startup
var debugConfig DebugConfig

func SetDebugConfig(cfg DebugConfig) {
	debugConfig = cfg
}

func isDebug() bool {
	return debugConfig.IsDebug()
}

func writeJSON(w http.ResponseWriter, r *http.Request, ae APIError) {
	if ae.CorrelationID == "" {
		if rid := middleware.GetReqID(r.Context()); rid != "" {
			ae.CorrelationID = rid
		}
	}

	// Add debug information in debug mode
	if isDebug() {
		if ae.Details == nil {
			ae.Details = map[string]any{}
		}
		if ae.debugMsg != "" {
			ae.Details["debug"] = ae.debugMsg
		}
		if ae.stackTrace != "" {
			ae.Details["stack_trace"] = ae.stackTrace
		}
	}

	status := ae.httpStatus
	if status == 0 {
		status = http.StatusInternalServerError
	}
	render.Status(r, status)
	render.JSON(w, r, ae)
}
