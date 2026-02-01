package handlers

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/api/internal/auth"
	"github.com/aatuh/recsys-suite/api/internal/experiments"
	"github.com/aatuh/recsys-suite/api/internal/exposure"
	"github.com/aatuh/recsys-suite/api/internal/http/mapper"
	"github.com/aatuh/recsys-suite/api/internal/http/problem"
	appmetrics "github.com/aatuh/recsys-suite/api/internal/metrics"
	"github.com/aatuh/recsys-suite/api/internal/services/recsysvc"
	"github.com/aatuh/recsys-suite/api/internal/validation"
	endpointspec "github.com/aatuh/recsys-suite/api/src/specs/endpoints"
	"github.com/aatuh/recsys-suite/api/src/specs/types"

	"github.com/aatuh/api-toolkit/authorization"
	jsonmw "github.com/aatuh/api-toolkit/middleware/json"
	"github.com/aatuh/api-toolkit/ports"
	"github.com/aatuh/api-toolkit/response_writer"
)

const algoVersionStub = "recsys-algo@stub"

// RecsysHandler exposes recommendation endpoints.
type RecsysHandler struct {
	Svc                 *recsysvc.Service
	Logger              ports.Logger
	Validator           ports.Validator
	OverloadRetryAfter  time.Duration
	ExposureLogger      exposure.Logger
	ExposureHasher      exposure.Hasher
	ExperimentAssigner  experiments.Assigner
	ExplainMaxItems     int
	ExplainRequireAdmin bool
	AdminRole           string
}

// HandlerOption configures the handler.
type HandlerOption func(*RecsysHandler)

// WithOverloadRetryAfter sets the Retry-After header for overload responses.
func WithOverloadRetryAfter(d time.Duration) HandlerOption {
	return func(h *RecsysHandler) {
		h.OverloadRetryAfter = d
	}
}

// WithExposureLogger enables exposure logging.
func WithExposureLogger(logger exposure.Logger, hasher exposure.Hasher) HandlerOption {
	return func(h *RecsysHandler) {
		h.ExposureLogger = logger
		h.ExposureHasher = hasher
	}
}

// WithExperimentAssigner configures the experiment assignment hook.
func WithExperimentAssigner(assigner experiments.Assigner) HandlerOption {
	return func(h *RecsysHandler) {
		h.ExperimentAssigner = assigner
	}
}

// WithExplainControls configures explain/trace safeguards.
func WithExplainControls(maxItems int, requireAdmin bool, adminRole string) HandlerOption {
	return func(h *RecsysHandler) {
		h.ExplainMaxItems = maxItems
		h.ExplainRequireAdmin = requireAdmin
		h.AdminRole = adminRole
	}
}

// NewRecsysHandler constructs a new handler.
func NewRecsysHandler(s *recsysvc.Service, l ports.Logger, v ports.Validator, opts ...HandlerOption) *RecsysHandler {
	h := &RecsysHandler{Svc: s, Logger: l, Validator: v}
	for _, opt := range opts {
		if opt != nil {
			opt(h)
		}
	}
	return h
}

// RegisterRoutes mounts recsys endpoints on the router.
func (h *RecsysHandler) RegisterRoutes(r ports.HTTPRouter) {
	r.Post(endpointspec.Recommend, h.recommend)
	r.Post(endpointspec.RecommendValidate, h.validate)
	r.Post(endpointspec.Similar, h.similar)
}

// recommend handles POST /v1/recommend.
func (h *RecsysHandler) recommend(w http.ResponseWriter, r *http.Request) {
	var dto types.RecommendRequest
	if err := decodeStrictJSON(r, &dto); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
		return
	}
	if err := h.Validator.ValidateStruct(r.Context(), &dto); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
		return
	}

	norm, warnings, err := validation.NormalizeRecommendRequest(&dto)
	if err != nil {
		writeValidationError(w, r, err)
		return
	}
	if !h.enforceExplainControls(w, r, norm.Options, norm.K) {
		return
	}
	if !h.enforceExplainControls(w, r, norm.Options, norm.K) {
		return
	}
	if h.ExperimentAssigner != nil {
		norm.Experiment = h.ExperimentAssigner.Assign(norm.Experiment, norm.User)
	}

	items, svcWarnings, meta, err := h.Svc.Recommend(r.Context(), norm)
	if err != nil {
		if errors.Is(err, recsysvc.ErrOverloaded) {
			appmetrics.RecordBackpressureRejection()
			h.writeOverloaded(w, r)
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "RECSYS_INTERNAL", "internal error")
		return
	}

	allWarnings := append(warnings, svcWarnings...)
	resp := types.RecommendResponse{
		Items:    mapper.RecommendItemsDTO(items),
		Meta:     buildMeta(norm, r, len(items), meta),
		Warnings: mapper.WarningsDTO(allWarnings),
	}
	problem.SetRequestIDHeader(w, r)
	w.Header().Set("Cache-Control", "no-store")
	response_writer.WriteJSON(w, http.StatusOK, resp)
	h.logExposure(r, norm, items, meta)
}

// similar handles POST /v1/similar.
func (h *RecsysHandler) similar(w http.ResponseWriter, r *http.Request) {
	var dto types.SimilarRequest
	if err := decodeStrictJSON(r, &dto); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
		return
	}
	if err := h.Validator.ValidateStruct(r.Context(), &dto); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
		return
	}

	norm, warnings, err := validation.NormalizeSimilarRequest(&dto)
	if err != nil {
		writeValidationError(w, r, err)
		return
	}
	if !h.enforceExplainControls(w, r, norm.Options, norm.K) {
		return
	}

	items, svcWarnings, meta, err := h.Svc.Similar(r.Context(), norm)
	if err != nil {
		if errors.Is(err, recsysvc.ErrOverloaded) {
			appmetrics.RecordBackpressureRejection()
			h.writeOverloaded(w, r)
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "RECSYS_INTERNAL", "internal error")
		return
	}

	allWarnings := append(warnings, svcWarnings...)
	resp := types.RecommendResponse{
		Items:    mapper.RecommendItemsDTO(items),
		Meta:     buildMetaFromSimilar(norm, r, len(items), meta),
		Warnings: mapper.WarningsDTO(allWarnings),
	}
	problem.SetRequestIDHeader(w, r)
	w.Header().Set("Cache-Control", "no-store")
	response_writer.WriteJSON(w, http.StatusOK, resp)
	h.logExposureFromSimilar(r, norm, items, meta)
}

func (h *RecsysHandler) writeOverloaded(w http.ResponseWriter, r *http.Request) {
	if h != nil && h.OverloadRetryAfter > 0 {
		seconds := int64(h.OverloadRetryAfter.Seconds())
		if seconds <= 0 {
			seconds = 1
		}
		w.Header().Set("Retry-After", strconv.FormatInt(seconds, 10))
	}
	writeProblem(w, r, http.StatusServiceUnavailable, "RECSYS_OVERLOADED", "service overloaded")
}

func (h *RecsysHandler) enforceExplainControls(w http.ResponseWriter, r *http.Request, opts recsysvc.Options, k int) bool {
	if h == nil {
		return true
	}
	explain := strings.ToLower(strings.TrimSpace(opts.Explain))
	if explain != "none" && h.ExplainMaxItems > 0 && k > h.ExplainMaxItems {
		writeProblem(w, r, http.StatusUnprocessableEntity, "RECSYS_EXPLAIN_TOO_LARGE", "explain payload too large")
		return false
	}
	if h.ExplainRequireAdmin && (opts.IncludeTrace || explain == "full") {
		if !h.hasAdminRole(r) {
			writeProblem(w, r, http.StatusForbidden, "RECSYS_FORBIDDEN", "insufficient scope")
			return false
		}
	}
	return true
}

func (h *RecsysHandler) hasAdminRole(r *http.Request) bool {
	if h == nil {
		return false
	}
	required := strings.TrimSpace(h.AdminRole)
	if required == "" {
		return true
	}
	roles := auth.RolesFromContext(r.Context())
	for _, role := range roles {
		if strings.EqualFold(strings.TrimSpace(role), required) {
			return true
		}
	}
	return false
}

// validate handles POST /v1/recommend/validate.
func (h *RecsysHandler) validate(w http.ResponseWriter, r *http.Request) {
	var dto types.RecommendRequest
	if err := decodeStrictJSON(r, &dto); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
		return
	}
	if err := h.Validator.ValidateStruct(r.Context(), &dto); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "RECSYS_INVALID_REQUEST", err.Error())
		return
	}

	norm, warnings, err := validation.NormalizeRecommendRequest(&dto)
	if err != nil {
		writeValidationError(w, r, err)
		return
	}

	resp := types.ValidateResponse{
		NormalizedRequest: mapper.NormalizedRecommendRequestDTO(norm),
		Warnings:          mapper.WarningsDTO(warnings),
		Meta:              buildValidateMeta(norm, r),
	}
	problem.SetRequestIDHeader(w, r)
	response_writer.WriteJSON(w, http.StatusOK, resp)
}

func decodeStrictJSON(r *http.Request, dst any) error {
	dec, err := jsonmw.StrictDecoder(r)
	if err != nil {
		return err
	}
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return errors.New("unexpected trailing data")
}

func buildValidateMeta(req recsysvc.RecommendRequest, r *http.Request) types.ResponseMeta {
	return buildMetaWithVersions(req.Surface, req.Segment, r, 0, recsysvc.ResponseMeta{})
}

func buildMeta(req recsysvc.RecommendRequest, r *http.Request, returned int, meta recsysvc.ResponseMeta) types.ResponseMeta {
	return buildMetaWithVersions(req.Surface, req.Segment, r, returned, meta)
}

func buildMetaFromSimilar(req recsysvc.SimilarRequest, r *http.Request, returned int, meta recsysvc.ResponseMeta) types.ResponseMeta {
	return buildMetaWithVersions(req.Surface, req.Segment, r, returned, meta)
}

func buildMetaWithVersions(surface, segment string, r *http.Request, returned int, meta recsysvc.ResponseMeta) types.ResponseMeta {
	algoVersion := meta.AlgoVersion
	if strings.TrimSpace(algoVersion) == "" {
		algoVersion = algoVersionStub
	}
	respMeta := types.ResponseMeta{
		TenantID:      tenantIDFromRequest(r),
		Surface:       surface,
		Segment:       segment,
		AlgoVersion:   algoVersion,
		ConfigVersion: meta.ConfigVersion,
		RulesVersion:  meta.RulesVersion,
		RequestID:     problem.RequestID(r),
	}
	if returned > 0 {
		respMeta.Counts = map[string]int{"returned": returned}
	}
	return respMeta
}

func (h *RecsysHandler) logExposure(r *http.Request, req recsysvc.RecommendRequest, items []recsysvc.Item, meta recsysvc.ResponseMeta) {
	if h == nil || h.ExposureLogger == nil {
		return
	}
	event := exposure.Event{
		RequestID:     problem.RequestID(r),
		TenantID:      tenantIDFromRequest(r),
		Surface:       req.Surface,
		Segment:       req.Segment,
		AlgoVersion:   meta.AlgoVersion,
		ConfigVersion: meta.ConfigVersion,
		RulesVersion:  meta.RulesVersion,
		Experiment:    mapExperiment(req.Experiment),
		Context:       mapContext(req.Context),
		Subject:       h.ExposureHasher.Subject(req.User.UserID, req.User.AnonymousID, req.User.SessionID),
		Items:         mapExposureItems(items),
	}
	if err := h.ExposureLogger.Log(r.Context(), event); err != nil && h.Logger != nil {
		h.Logger.Error("exposure log failed", "err", err)
	}
}

func (h *RecsysHandler) logExposureFromSimilar(r *http.Request, req recsysvc.SimilarRequest, items []recsysvc.Item, meta recsysvc.ResponseMeta) {
	if h == nil || h.ExposureLogger == nil {
		return
	}
	event := exposure.Event{
		RequestID:     problem.RequestID(r),
		TenantID:      tenantIDFromRequest(r),
		Surface:       req.Surface,
		Segment:       req.Segment,
		AlgoVersion:   meta.AlgoVersion,
		ConfigVersion: meta.ConfigVersion,
		RulesVersion:  meta.RulesVersion,
		Items:         mapExposureItems(items),
	}
	if err := h.ExposureLogger.Log(r.Context(), event); err != nil && h.Logger != nil {
		h.Logger.Error("exposure log failed", "err", err)
	}
}

func mapExposureItems(items []recsysvc.Item) []exposure.Item {
	if len(items) == 0 {
		return nil
	}
	out := make([]exposure.Item, len(items))
	for i, item := range items {
		out[i] = exposure.Item{
			ItemID: item.ItemID,
			Rank:   item.Rank,
			Score:  item.Score,
		}
	}
	return out
}

func mapExperiment(exp *recsysvc.Experiment) *exposure.Experiment {
	if exp == nil {
		return nil
	}
	if strings.TrimSpace(exp.ID) == "" && strings.TrimSpace(exp.Variant) == "" {
		return nil
	}
	return &exposure.Experiment{ID: exp.ID, Variant: exp.Variant}
}

func mapContext(ctx *recsysvc.RequestContext) *exposure.Context {
	if ctx == nil {
		return nil
	}
	if strings.TrimSpace(ctx.Locale) == "" && strings.TrimSpace(ctx.Device) == "" && strings.TrimSpace(ctx.Country) == "" && strings.TrimSpace(ctx.Now) == "" {
		return nil
	}
	return &exposure.Context{
		Locale:  ctx.Locale,
		Device:  ctx.Device,
		Country: ctx.Country,
		Now:     ctx.Now,
	}
}

func tenantIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if tenant, ok := authorization.TenantIDFromContext(r.Context()); ok {
		return tenant
	}
	return strings.TrimSpace(r.Header.Get("X-Org-Id"))
}
