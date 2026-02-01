package handlers

import (
	"net/http"

	"github.com/aatuh/recsys-suite/api/internal/http/problem"
	"github.com/aatuh/recsys-suite/api/internal/validation"
)

func writeValidationError(w http.ResponseWriter, r *http.Request, err error) {
	if verr, ok := err.(validation.Error); ok {
		writeProblem(w, r, verr.Status, verr.Code, verr.Message)
		return
	}
	writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, detail string) {
	problem.Write(w, r, status, code, detail)
}
