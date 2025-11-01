package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"recsys/internal/http/common"
	"recsys/internal/store"
	handlerstypes "recsys/specs/types"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// AuditHandler serves decision-trace audit endpoints.
type AuditHandler struct {
	store      *store.Store
	defaultOrg uuid.UUID
}

func NewAuditHandler(st *store.Store, defaultOrg uuid.UUID) *AuditHandler {
	return &AuditHandler{store: st, defaultOrg: defaultOrg}
}

// AuditDecisionsList returns recent decision traces metadata.
// @Summary      List decision traces with optional filters
// @Tags         audit
// @Produce      json
// @Param        namespace query string true "Namespace"
// @Param        from query string false "From timestamp (RFC3339)"
// @Param        to query string false "To timestamp (RFC3339)"
// @Param        user_hash query string false "User hash"
// @Param        request_id query string false "Request ID"
// @Param        limit query int false "Limit"
// @Success      200 {object} types.AuditDecisionListResponse
// @Failure      400 {object} common.APIError
// @Router       /v1/audit/decisions [get]
func (h *AuditHandler) AuditDecisionsList(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		common.BadRequest(w, r, "missing_namespace", "namespace is required", nil)
		return
	}

	orgID := orgIDFromHeader(r, h.defaultOrg)

	filter, err := buildDecisionFilter(
		r.URL.Query().Get("from"),
		r.URL.Query().Get("to"),
		r.URL.Query().Get("user_hash"),
		r.URL.Query().Get("request_id"),
		r.URL.Query().Get("limit"),
	)
	if err != nil {
		common.BadRequest(w, r, err.Error(), "invalid query parameter", nil)
		return
	}

	resp, err := h.listDecisions(r.Context(), orgID, ns, filter)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// AuditDecisionsSearch accepts a JSON payload to filter decision traces.
// @Summary      Search decision traces with advanced filters
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        payload body types.AuditDecisionsSearchRequest true "Search request"
// @Success      200 {object} types.AuditDecisionListResponse
// @Failure      400 {object} common.APIError
// @Router       /v1/audit/search [post]
func (h *AuditHandler) AuditDecisionsSearch(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.AuditDecisionsSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, r, "invalid_payload", "request body must be valid JSON", nil)
		return
	}
	if req.Namespace == "" {
		common.BadRequest(w, r, "missing_namespace", "namespace is required", nil)
		return
	}

	limitStr := ""
	if req.Limit > 0 {
		limitStr = strconv.Itoa(req.Limit)
	}
	filter, err := buildDecisionFilter(req.From, req.To, req.UserHash, req.RequestID, limitStr)
	if err != nil {
		common.BadRequest(w, r, err.Error(), "invalid filter", nil)
		return
	}
	filter.Limit = req.Limit

	resp, err := h.listDecisions(r.Context(), orgIDFromHeader(r, h.defaultOrg), req.Namespace, filter)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// AuditDecisionGet returns the full trace for a decision id.
// @Summary      Get full decision trace by ID
// @Tags         audit
// @Produce      json
// @Param        decision_id path string true "Decision ID"
// @Success      200 {object} types.AuditDecisionDetail
// @Failure      400 {object} common.APIError
// @Failure      404 {object} common.APIError
// @Router       /v1/audit/decisions/{decision_id} [get]
func (h *AuditHandler) AuditDecisionGet(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "decision_id")
	decisionID, err := uuid.Parse(idStr)
	if err != nil {
		common.BadRequest(w, r, "invalid_decision_id", "decision_id must be a UUID", nil)
		return
	}

	orgID := orgIDFromHeader(r, h.defaultOrg)

	rec, err := h.store.GetDecisionTrace(r.Context(), orgID, decisionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			common.HttpError(w, r, err, http.StatusNotFound)
			return
		}
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	detail, err := convertDecisionRecord(*rec)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(detail)
}

func (h *AuditHandler) listDecisions(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	filter store.DecisionTraceFilter,
) (handlerstypes.AuditDecisionListResponse, error) {
	records, err := h.store.ListDecisionTraces(ctx, orgID, ns, filter)
	if err != nil {
		return handlerstypes.AuditDecisionListResponse{}, err
	}
	resp := handlerstypes.AuditDecisionListResponse{
		Decisions: make([]handlerstypes.AuditDecisionSummary, 0, len(records)),
	}
	for _, rec := range records {
		summary, err := convertDecisionSummary(rec)
		if err != nil {
			return handlerstypes.AuditDecisionListResponse{}, err
		}
		resp.Decisions = append(resp.Decisions, summary)
	}
	return resp, nil
}

func buildDecisionFilter(
	fromStr, toStr, userHash, requestID, limitStr string,
) (store.DecisionTraceFilter, error) {
	filter := store.DecisionTraceFilter{
		UserHash:  strings.TrimSpace(userHash),
		RequestID: strings.TrimSpace(requestID),
	}
	from, err := parseTimeString(fromStr)
	if err != nil {
		return filter, fmt.Errorf("invalid_from: %w", err)
	}
	filter.From = from
	to, err := parseTimeString(toStr)
	if err != nil {
		return filter, fmt.Errorf("invalid_to: %w", err)
	}
	filter.To = to
	if strings.TrimSpace(limitStr) != "" {
		lim, err := strconv.Atoi(strings.TrimSpace(limitStr))
		if err != nil {
			return filter, fmt.Errorf("invalid_limit: %w", err)
		}
		filter.Limit = lim
	}
	return filter, nil
}

func parseTimeString(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &ts, nil
}

func convertDecisionSummary(src store.DecisionTraceSummary) (handlerstypes.AuditDecisionSummary, error) {
	summary := handlerstypes.AuditDecisionSummary{
		DecisionID: src.DecisionID.String(),
		Namespace:  src.Namespace,
		Ts:         src.Timestamp,
	}
	if src.Surface != nil {
		summary.Surface = *src.Surface
	}
	if src.RequestID != nil {
		summary.RequestID = *src.RequestID
	}
	if src.UserHash != nil {
		summary.UserHash = *src.UserHash
	}
	if src.K != nil {
		summary.K = src.K
	}
	if len(src.FinalItemsJSON) > 0 {
		if err := json.Unmarshal(src.FinalItemsJSON, &summary.FinalItems); err != nil {
			return handlerstypes.AuditDecisionSummary{}, err
		}
	}
	if len(src.ExtrasJSON) > 0 {
		var extras map[string]any
		if err := json.Unmarshal(src.ExtrasJSON, &extras); err != nil {
			return handlerstypes.AuditDecisionSummary{}, err
		}
		summary.Extras = extras
	}
	return summary, nil
}

func convertDecisionRecord(src store.DecisionTraceRecord) (handlerstypes.AuditDecisionDetail, error) {
	detail := handlerstypes.AuditDecisionDetail{
		DecisionID: src.DecisionID.String(),
		OrgID:      src.OrgID.String(),
		Namespace:  src.Namespace,
		Ts:         src.Timestamp,
	}
	if src.Surface != nil {
		detail.Surface = *src.Surface
	}
	if src.RequestID != nil {
		detail.RequestID = *src.RequestID
	}
	if src.UserHash != nil {
		detail.UserHash = *src.UserHash
	}
	if src.K != nil {
		detail.K = src.K
	}
	if len(src.ConstraintsJSON) > 0 {
		var constraints handlerstypes.AuditTraceConstraints
		if err := json.Unmarshal(src.ConstraintsJSON, &constraints); err != nil {
			return handlerstypes.AuditDecisionDetail{}, err
		}
		detail.Constraints = &constraints
	}
	if len(src.EffectiveConfigJSON) > 0 {
		if err := json.Unmarshal(src.EffectiveConfigJSON, &detail.Config); err != nil {
			return handlerstypes.AuditDecisionDetail{}, err
		}
	}
	if len(src.BanditJSON) > 0 {
		var bandit handlerstypes.AuditTraceBandit
		if err := json.Unmarshal(src.BanditJSON, &bandit); err != nil {
			return handlerstypes.AuditDecisionDetail{}, err
		}
		detail.Bandit = &bandit
	}
	if len(src.CandidatesPreJSON) > 0 {
		if err := json.Unmarshal(src.CandidatesPreJSON, &detail.Candidates); err != nil {
			return handlerstypes.AuditDecisionDetail{}, err
		}
	}
	if len(src.FinalItemsJSON) > 0 {
		if err := json.Unmarshal(src.FinalItemsJSON, &detail.FinalItems); err != nil {
			return handlerstypes.AuditDecisionDetail{}, err
		}
	}
	if len(src.MMRInfoJSON) > 0 {
		if err := json.Unmarshal(src.MMRInfoJSON, &detail.MMR); err != nil {
			return handlerstypes.AuditDecisionDetail{}, err
		}
	}
	if len(src.CapsJSON) > 0 {
		var caps map[string]handlerstypes.AuditTraceCap
		if err := json.Unmarshal(src.CapsJSON, &caps); err != nil {
			return handlerstypes.AuditDecisionDetail{}, err
		}
		detail.Caps = caps
	}
	if len(src.ExtrasJSON) > 0 {
		var extras map[string]any
		if err := json.Unmarshal(src.ExtrasJSON, &extras); err != nil {
			return handlerstypes.AuditDecisionDetail{}, err
		}
		detail.Extras = extras
	}
	return detail, nil
}
