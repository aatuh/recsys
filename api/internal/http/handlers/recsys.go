package handlers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"recsys/internal/http/mapper"
	"recsys/internal/http/problem"
	"recsys/internal/services/recsysvc"
	"recsys/internal/validation"
	endpointspec "recsys/src/specs/endpoints"
	"recsys/src/specs/types"

	"github.com/aatuh/api-toolkit/authorization"
	jsonmw "github.com/aatuh/api-toolkit/middleware/json"
	"github.com/aatuh/api-toolkit/ports"
	"github.com/aatuh/api-toolkit/response_writer"
)

const algoVersionStub = "recsys-algo@stub"

// RecsysHandler exposes recommendation endpoints.
type RecsysHandler struct {
	Svc       *recsysvc.Service
	Logger    ports.Logger
	Validator ports.Validator
}

// NewRecsysHandler constructs a new handler.
func NewRecsysHandler(s *recsysvc.Service, l ports.Logger, v ports.Validator) *RecsysHandler {
	return &RecsysHandler{Svc: s, Logger: l, Validator: v}
}

// RegisterRoutes mounts recsys endpoints on the router.
func (h *RecsysHandler) RegisterRoutes(r ports.HTTPRouter) {
	r.Post(endpointspec.Recommend, h.recommend)
	r.Post(endpointspec.RecommendValidate, h.validate)
	r.Post(endpointspec.Similar, h.similar)
}

// recommend handles POST /v1/recommend.
// @Summary Recommend items
// @Description Return ranked recommendations
// @Tags Recsys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body types.RecommendRequest true "Recommend payload"
// @Success 200 {object} types.RecommendResponse
// @Failure 400 {object} types.Problem
// @Failure 422 {object} types.Problem
// @Router /v1/recommend [post]
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

	items, svcWarnings, err := h.Svc.Recommend(r.Context(), norm)
	if err != nil {
		writeProblem(w, r, http.StatusInternalServerError, "RECSYS_INTERNAL", "internal error")
		return
	}

	allWarnings := append(warnings, svcWarnings...)
	resp := types.RecommendResponse{
		Items:    mapper.RecommendItemsDTO(items),
		Meta:     buildMeta(norm, r, len(items)),
		Warnings: mapper.WarningsDTO(allWarnings),
	}
	problem.SetRequestIDHeader(w, r)
	w.Header().Set("Cache-Control", "no-store")
	response_writer.WriteJSON(w, http.StatusOK, resp)
}

// similar handles POST /v1/similar.
// @Summary Similar items
// @Description Return similar items for an item_id
// @Tags Recsys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body types.SimilarRequest true "Similar payload"
// @Success 200 {object} types.RecommendResponse
// @Failure 400 {object} types.Problem
// @Failure 422 {object} types.Problem
// @Router /v1/similar [post]
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

	items, svcWarnings, err := h.Svc.Similar(r.Context(), norm)
	if err != nil {
		writeProblem(w, r, http.StatusInternalServerError, "RECSYS_INTERNAL", "internal error")
		return
	}

	allWarnings := append(warnings, svcWarnings...)
	resp := types.RecommendResponse{
		Items:    mapper.RecommendItemsDTO(items),
		Meta:     buildMetaFromSimilar(norm, r, len(items)),
		Warnings: mapper.WarningsDTO(allWarnings),
	}
	problem.SetRequestIDHeader(w, r)
	w.Header().Set("Cache-Control", "no-store")
	response_writer.WriteJSON(w, http.StatusOK, resp)
}

// validate handles POST /v1/recommend/validate.
// @Summary Validate recommend request
// @Description Validate and normalize a recommend request
// @Tags Recsys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body types.RecommendRequest true "Recommend payload"
// @Success 200 {object} types.ValidateResponse
// @Failure 400 {object} types.Problem
// @Failure 422 {object} types.Problem
// @Router /v1/recommend/validate [post]
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
		Meta:              buildMeta(norm, r, 0),
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

func buildMeta(req recsysvc.RecommendRequest, r *http.Request, returned int) types.ResponseMeta {
	meta := types.ResponseMeta{
		TenantID:    tenantIDFromRequest(r),
		Surface:     req.Surface,
		Segment:     req.Segment,
		AlgoVersion: algoVersionStub,
		RequestID:   problem.RequestID(r),
	}
	if returned > 0 {
		meta.Counts = map[string]int{"returned": returned}
	}
	return meta
}

func buildMetaFromSimilar(req recsysvc.SimilarRequest, r *http.Request, returned int) types.ResponseMeta {
	meta := types.ResponseMeta{
		TenantID:    tenantIDFromRequest(r),
		Surface:     req.Surface,
		Segment:     req.Segment,
		AlgoVersion: algoVersionStub,
		RequestID:   problem.RequestID(r),
	}
	if returned > 0 {
		meta.Counts = map[string]int{"returned": returned}
	}
	return meta
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
