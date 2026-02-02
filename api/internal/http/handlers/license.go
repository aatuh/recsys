package handlers

import (
	"net/http"

	"github.com/aatuh/recsys-suite/api/internal/http/mapper"
	"github.com/aatuh/recsys-suite/api/internal/license"
	endpointspec "github.com/aatuh/recsys-suite/api/src/specs/endpoints"

	"github.com/aatuh/api-toolkit/v2/ports"
	"github.com/aatuh/api-toolkit/v2/response_writer"
)

// LicenseHandler exposes license status endpoint.
type LicenseHandler struct {
	Manager *license.Manager
	Logger  ports.Logger
}

// NewLicenseHandler constructs a new license handler.
func NewLicenseHandler(manager *license.Manager, log ports.Logger) *LicenseHandler {
	return &LicenseHandler{Manager: manager, Logger: log}
}

// RegisterRoutes mounts license endpoints on the router.
func (h *LicenseHandler) RegisterRoutes(r ports.HTTPRouter) {
	if r == nil {
		return
	}
	r.Get(endpointspec.LicenseStatus, h.getStatus)
}

func (h *LicenseHandler) getStatus(w http.ResponseWriter, r *http.Request) {
	info, err := h.status(r)
	if err != nil && h.Logger != nil {
		h.Logger.Warn("license status lookup failed", "err", err.Error())
	}
	w.Header().Set("Cache-Control", "no-store")
	response_writer.WriteJSON(w, http.StatusOK, mapper.LicenseStatusDTO(info))
}

func (h *LicenseHandler) status(r *http.Request) (license.Info, error) {
	if h == nil || h.Manager == nil {
		return license.Info{Status: license.StatusUnknown}, nil
	}
	return h.Manager.Status(r.Context())
}
