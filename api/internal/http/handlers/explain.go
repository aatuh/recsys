package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"recsys/internal/explain"
	"recsys/internal/http/common"
	handlerstypes "recsys/specs/types"

	"github.com/google/uuid"
)

// ExplainHandler manages LLM explain endpoints.
type ExplainHandler struct {
	service    *explain.Service
	defaultOrg uuid.UUID
}

func NewExplainHandler(service *explain.Service, defaultOrg uuid.UUID) *ExplainHandler {
	return &ExplainHandler{service: service, defaultOrg: defaultOrg}
}

// ExplainLLM godoc
// @Summary      Generate RCA explanation via LLM
// @Tags         explain
// @Accept       json
// @Produce      json
// @Param        payload body types.ExplainLLMRequest true "Explain request"
// @Success      200 {object} types.ExplainLLMResponse
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/explain/llm [post]
func (h *ExplainHandler) ExplainLLM(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		common.ServiceUnavailable(w, r)
		return
	}

	var req handlerstypes.ExplainLLMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}

	req.TargetType = strings.TrimSpace(req.TargetType)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.Namespace = strings.TrimSpace(req.Namespace)
	req.Surface = strings.TrimSpace(req.Surface)
	req.SegmentID = strings.TrimSpace(req.SegmentID)

	if req.TargetType == "" || req.TargetID == "" || req.Namespace == "" || req.Surface == "" || req.From == "" || req.To == "" {
		common.BadRequest(w, r, "invalid_request", "target_type, target_id, namespace, surface, from, and to are required", nil)
		return
	}

	from, err := time.Parse(time.RFC3339, req.From)
	if err != nil {
		common.BadRequest(w, r, "invalid_from", "from must be RFC3339", nil)
		return
	}
	to, err := time.Parse(time.RFC3339, req.To)
	if err != nil {
		common.BadRequest(w, r, "invalid_to", "to must be RFC3339", nil)
		return
	}
	if !to.After(from) {
		common.BadRequest(w, r, "invalid_range", "to must be after from", nil)
		return
	}

	orgID := orgIDFromHeader(r, h.defaultOrg)

	explainReq := explain.Request{
		OrgID:      orgID.String(),
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Namespace:  req.Namespace,
		Surface:    req.Surface,
		SegmentID:  req.SegmentID,
		From:       from,
		To:         to,
		Question:   req.Question,
	}

	result, err := h.service.Explain(r.Context(), orgID, explainReq)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	factsMap := map[string]any{}
	rawFacts, _ := json.Marshal(result.Facts)
	_ = json.Unmarshal(rawFacts, &factsMap)

	resp := handlerstypes.ExplainLLMResponse{
		Markdown: result.Markdown,
		Cache:    map[bool]string{true: "hit", false: "miss"}[result.CacheHit],
		Model:    result.Model,
		Facts:    factsMap,
		Warnings: result.Warnings,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Explain-Cache", resp.Cache)
	_ = json.NewEncoder(w).Encode(resp)
}
