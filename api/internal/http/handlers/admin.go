package handlers

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/api/internal/admin"
	"github.com/aatuh/recsys-suite/api/internal/auth"
	"github.com/aatuh/recsys-suite/api/internal/http/mapper"
	"github.com/aatuh/recsys-suite/api/internal/http/problem"
	"github.com/aatuh/recsys-suite/api/internal/services/adminsvc"
	"github.com/aatuh/recsys-suite/api/internal/validation"
	endpointspec "github.com/aatuh/recsys-suite/api/src/specs/endpoints"
	"github.com/aatuh/recsys-suite/api/src/specs/types"

	"github.com/aatuh/api-toolkit/contrib/v2/adapters/chi"
	"github.com/aatuh/api-toolkit/v2/ports"
	"github.com/aatuh/api-toolkit/v2/response_writer"
)

// AdminHandler exposes admin/control-plane endpoints.
type AdminHandler struct {
	Svc       *adminsvc.Service
	Logger    ports.Logger
	Validator ports.Validator
}

// NewAdminHandler constructs a new admin handler.
func NewAdminHandler(s *adminsvc.Service, l ports.Logger, v ports.Validator) *AdminHandler {
	return &AdminHandler{Svc: s, Logger: l, Validator: v}
}

// Routes returns a router with admin endpoints.
func (h *AdminHandler) Routes() ports.HTTPRouter {
	r := chi.New()
	r.Get(endpointspec.AdminTenantConfig, h.getConfig)
	r.Put(endpointspec.AdminTenantConfig, h.putConfig)
	r.Get(endpointspec.AdminTenantRules, h.getRules)
	r.Put(endpointspec.AdminTenantRules, h.putRules)
	r.Post(endpointspec.AdminTenantInvalidate, h.invalidateCache)
	r.Get(endpointspec.AdminTenantAudit, h.getAudit)
	return r
}

// getConfig handles GET /v1/admin/tenants/{tenant_id}/config.
func (h *AdminHandler) getConfig(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(chi.URLParam(r, "tenant_id"))
	if tenantID == "" {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_TENANT", "tenant_id is required")
		return
	}
	cfg, err := h.Svc.GetTenantConfig(r.Context(), tenantID)
	if err != nil {
		h.writeAdminErr(w, r, err)
		return
	}
	w.Header().Set("ETag", cfg.Version)
	response_writer.WriteJSON(w, http.StatusOK, mapper.TenantConfigResponse(cfg))
}

// putConfig handles PUT /v1/admin/tenants/{tenant_id}/config.
func (h *AdminHandler) putConfig(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(chi.URLParam(r, "tenant_id"))
	if tenantID == "" {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_TENANT", "tenant_id is required")
		return
	}
	raw, err := readRawJSON(w, r)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
		return
	}
	if err := validation.ValidateConfigPayload(raw); err != nil {
		writeValidationError(w, r, err)
		return
	}
	ifMatch := r.Header.Get("If-Match")
	actor := actorFromRequest(r)
	meta := requestMeta(r)
	cfg, err := h.Svc.UpdateTenantConfig(r.Context(), tenantID, raw, ifMatch, actor, meta)
	if err != nil {
		h.writeAdminErr(w, r, err)
		return
	}
	w.Header().Set("ETag", cfg.Version)
	response_writer.WriteJSON(w, http.StatusOK, mapper.TenantConfigResponse(cfg))
}

// getRules handles GET /v1/admin/tenants/{tenant_id}/rules.
func (h *AdminHandler) getRules(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(chi.URLParam(r, "tenant_id"))
	if tenantID == "" {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_TENANT", "tenant_id is required")
		return
	}
	rules, err := h.Svc.GetTenantRules(r.Context(), tenantID)
	if err != nil {
		h.writeAdminErr(w, r, err)
		return
	}
	w.Header().Set("ETag", rules.Version)
	response_writer.WriteJSON(w, http.StatusOK, mapper.TenantRulesResponse(rules))
}

// putRules handles PUT /v1/admin/tenants/{tenant_id}/rules.
func (h *AdminHandler) putRules(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(chi.URLParam(r, "tenant_id"))
	if tenantID == "" {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_TENANT", "tenant_id is required")
		return
	}
	raw, err := readRawJSON(w, r)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
		return
	}
	if err := validation.ValidateRulesPayload(raw); err != nil {
		writeValidationError(w, r, err)
		return
	}
	ifMatch := r.Header.Get("If-Match")
	actor := actorFromRequest(r)
	meta := requestMeta(r)
	rules, err := h.Svc.UpdateTenantRules(r.Context(), tenantID, raw, ifMatch, actor, meta)
	if err != nil {
		h.writeAdminErr(w, r, err)
		return
	}
	w.Header().Set("ETag", rules.Version)
	response_writer.WriteJSON(w, http.StatusOK, mapper.TenantRulesResponse(rules))
}

// invalidateCache handles POST /v1/admin/tenants/{tenant_id}/cache/invalidate.
func (h *AdminHandler) invalidateCache(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(chi.URLParam(r, "tenant_id"))
	if tenantID == "" {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_TENANT", "tenant_id is required")
		return
	}
	var dto types.CacheInvalidateRequest
	if err := decodeStrictJSON(r, &dto); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
		return
	}
	if err := validation.ValidateCacheInvalidate(&dto); err != nil {
		writeValidationError(w, r, err)
		return
	}
	actor := actorFromRequest(r)
	meta := requestMeta(r)
	result, err := h.Svc.InvalidateCache(r.Context(), tenantID, adminsvc.CacheInvalidateRequest{
		Targets: dto.Targets,
		Surface: dto.Surface,
	}, actor, meta)
	if err != nil {
		h.writeAdminErr(w, r, err)
		return
	}
	response_writer.WriteJSON(w, http.StatusOK, mapper.CacheInvalidateResponse(result))
}

// getAudit handles GET /v1/admin/tenants/{tenant_id}/audit.
func (h *AdminHandler) getAudit(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(chi.URLParam(r, "tenant_id"))
	if tenantID == "" {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_TENANT", "tenant_id is required")
		return
	}
	query := r.URL.Query()
	limit, err := parseLimit(query.Get("limit"), adminsvc.DefaultAuditLimit, adminsvc.MaxAuditLimit)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
		return
	}
	before, err := parseTime(query.Get("before"))
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
		return
	}
	beforeID, err := parseBeforeID(query.Get("before_id"))
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
		return
	}
	log, err := h.Svc.ListAuditLog(r.Context(), tenantID, adminsvc.AuditQuery{
		Limit:    limit,
		Before:   before,
		BeforeID: beforeID,
	})
	if err != nil {
		h.writeAdminErr(w, r, err)
		return
	}
	response_writer.WriteJSON(w, http.StatusOK, mapper.AuditLogResponse(log))
}

func (h *AdminHandler) writeAdminErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, admin.ErrTenantNotFound):
		writeProblem(w, r, http.StatusNotFound, "RECSYS_TENANT_NOT_FOUND", "tenant not found")
	case errors.Is(err, admin.ErrConfigNotFound):
		writeProblem(w, r, http.StatusNotFound, "RECSYS_CONFIG_NOT_FOUND", "tenant config not found")
	case errors.Is(err, admin.ErrRulesNotFound):
		writeProblem(w, r, http.StatusNotFound, "RECSYS_RULES_NOT_FOUND", "tenant rules not found")
	case errors.Is(err, admin.ErrVersionMismatch):
		writeProblem(w, r, http.StatusConflict, "RECSYS_VERSION_MISMATCH", "version mismatch")
	default:
		writeProblem(w, r, http.StatusInternalServerError, "RECSYS_INTERNAL", "internal error")
		if h.Logger != nil {
			h.Logger.Error("admin request failed", "err", err)
		}
	}
}

func actorFromRequest(r *http.Request) adminsvc.Actor {
	info, _ := auth.FromContext(r.Context())
	return adminsvc.Actor{ID: info.UserID, Type: "user"}
}

func requestMeta(r *http.Request) adminsvc.RequestMeta {
	meta := adminsvc.RequestMeta{
		RequestID: problem.RequestID(r),
		UserAgent: "",
		IP:        nil,
	}
	if r == nil {
		return meta
	}
	meta.UserAgent = r.UserAgent()
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		meta.IP = net.ParseIP(host)
	} else if ip := net.ParseIP(r.RemoteAddr); ip != nil {
		meta.IP = ip
	}
	return meta
}

func readRawJSON(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if r == nil || r.Body == nil {
		return nil, io.EOF
	}
	const maxBody = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func parseLimit(raw string, def, max int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New("limit must be an integer")
	}
	if val <= 0 {
		return 0, errors.New("limit must be positive")
	}
	if val > max {
		return max, nil
	}
	return val, nil
}

func parseTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	if ts, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return ts, nil
	}
	if ts, err := time.Parse(time.RFC3339, raw); err == nil {
		return ts, nil
	}
	return time.Time{}, errors.New("before must be RFC3339 timestamp")
}

func parseBeforeID(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, errors.New("before_id must be an integer")
	}
	if val < 0 {
		return 0, errors.New("before_id must be non-negative")
	}
	return val, nil
}
