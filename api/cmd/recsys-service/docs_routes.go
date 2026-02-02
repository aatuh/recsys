package main

import (
	"net/http"

	"github.com/aatuh/recsys-suite/api/internal/http/problem"

	toolkitdocs "github.com/aatuh/api-toolkit/v2/endpoints/docs"
	"github.com/aatuh/api-toolkit/v2/ports"
)

func mountDocsRoutes(r ports.HTTPRouter, handler *toolkitdocs.Handler, openapiJSON, openapiYAML []byte) {
	if handler == nil || r == nil {
		return
	}
	paths := ports.DefaultDocsPaths()
	handler.RegisterCustomRoutes(r, ports.DocsPaths{
		HTML:    paths.HTML,
		Version: paths.Version,
		Info:    paths.Info,
	})
	r.Get(paths.OpenAPI, openAPIJSONHandler(openapiJSON))
	r.Get("/openapi.yaml", openAPIYAMLHandler(openapiYAML))
}

func openAPIJSONHandler(spec []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(spec) == 0 {
			problem.Write(w, r, http.StatusNotFound, "RECSYS_OPENAPI_NOT_FOUND", "openapi specification not found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(spec)
	}
}

func openAPIYAMLHandler(spec []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(spec) == 0 {
			problem.Write(w, r, http.StatusNotFound, "RECSYS_OPENAPI_NOT_FOUND", "openapi specification not found")
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(spec)
	}
}
