package common

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5/pgconn"
)

type APIError struct {
	Code          string         `json:"error"`
	Message       string         `json:"message,omitempty"`
	Details       map[string]any `json:"details,omitempty"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	httpStatus    int            `json:"-"`
	debugMsg      string         `json:"-"`
}

func NewAPIError(code, msg string, status int) APIError {
	return APIError{Code: code, Message: msg, httpStatus: status}
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
func ServiceUnavailable(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, NewAPIError("service_unavailable", "Service unavailable", http.StatusServiceUnavailable))
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
	if isDebug() && ae.debugMsg != "" {
		if ae.Details == nil {
			ae.Details = map[string]any{}
		}
		ae.Details["debug"] = ae.debugMsg
	}
	status := ae.httpStatus
	if status == 0 {
		status = http.StatusInternalServerError
	}
	render.Status(r, status)
	render.JSON(w, r, ae)
}
