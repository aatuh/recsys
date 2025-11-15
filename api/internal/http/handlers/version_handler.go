package handlers

import (
	"encoding/json"
	"net/http"

	"recsys/internal/version"
)

// VersionResponse documents the metadata exposed via /version.
type VersionResponse struct {
	GitCommit    string `json:"git_commit" example:"d34db33f"`
	BuildTime    string `json:"build_time" example:"2025-01-15T12:34:56Z"`
	ModelVersion string `json:"model_version" example:"popularity_v1"`
}

// NewVersionHandler returns a handler that reports build/runtime metadata.
//
// Version godoc
// @Summary      Describe the running build
// @Tags         meta
// @Produce      json
// @Success      200  {object}  handlers.VersionResponse
// @Router       /version [get]
func NewVersionHandler(info version.Info) http.HandlerFunc {
	payload := VersionResponse{
		GitCommit:    info.GitCommit,
		BuildTime:    info.BuildTime,
		ModelVersion: info.ModelVersion,
	}

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}
}
