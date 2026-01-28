package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aatuh/recsys-algo/rules"
	"recsys/internal/http/common"
	"recsys/internal/store"
	"recsys/internal/types"
	specstypes "recsys/specs/types"

	recmodel "github.com/aatuh/recsys-algo/model"
)

// RulesHandler exposes merchandising rule management endpoints.
type RulesHandler struct {
	store               *store.Store
	manager             *rules.Manager
	defaultOrg          uuid.UUID
	brandTagPrefixes    []string
	categoryTagPrefixes []string
}

// NewRulesHandler constructs a handler for merchandising rules.
func NewRulesHandler(st *store.Store, mgr *rules.Manager, defaultOrg uuid.UUID, brandPrefixes, categoryPrefixes []string) *RulesHandler {
	return &RulesHandler{
		store:               st,
		manager:             mgr,
		defaultOrg:          defaultOrg,
		brandTagPrefixes:    append([]string(nil), brandPrefixes...),
		categoryTagPrefixes: append([]string(nil), categoryPrefixes...),
	}
}

// RulesCreate godoc
// @Summary      Create a merchandising rule
// @Description  Create a new merchandising rule (BLOCK, PIN, BOOST)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        payload body types.RulePayload true "Rule payload"
// @Success      201 {object} types.RuleResponse
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/admin/rules [post]
func (h *RulesHandler) RulesCreate(w http.ResponseWriter, r *http.Request) {
	var payload specstypes.RulePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}

	input, err := parseRulePayload(payload)
	if err != nil {
		common.BadRequest(w, r, "invalid_rule", err.Error(), nil)
		return
	}

	orgID := orgIDFromHeader(r, h.defaultOrg)
	rule := types.Rule{
		RuleID:      uuid.New(),
		OrgID:       orgID,
		Namespace:   input.Namespace,
		Surface:     input.Surface,
		Name:        input.Name,
		Description: input.Description,
		Action:      input.Action,
		TargetType:  input.TargetType,
		TargetKey:   input.TargetKey,
		ItemIDs:     input.ItemIDs,
		BoostValue:  input.BoostValue,
		MaxPins:     input.MaxPins,
		SegmentID:   input.SegmentID,
		Priority:    input.Priority,
		Enabled:     input.Enabled,
		ValidFrom:   input.ValidFrom,
		ValidUntil:  input.ValidUntil,
	}

	created, err := h.store.CreateRule(r.Context(), rule)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	h.invalidateRuleCache(rule.OrgID, rule.Namespace, rule.Surface)

	resp := toRuleResponse(*created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// RulesUpdate godoc
// @Summary      Update a merchandising rule
// @Description  Update an existing merchandising rule
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        rule_id path string true "Rule ID"
// @Param        payload body types.RulePayload true "Rule payload"
// @Success      200 {object} types.RuleResponse
// @Failure      400 {object} common.APIError
// @Failure      404 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/admin/rules/{rule_id} [put]
func (h *RulesHandler) RulesUpdate(w http.ResponseWriter, r *http.Request) {
	ruleIDParam := chi.URLParam(r, "rule_id")
	ruleID, err := uuid.Parse(ruleIDParam)
	if err != nil {
		common.BadRequest(w, r, "invalid_rule_id", "rule_id must be a UUID", nil)
		return
	}

	var payload specstypes.RulePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	input, err := parseRulePayload(payload)
	if err != nil {
		common.BadRequest(w, r, "invalid_rule", err.Error(), nil)
		return
	}

	orgID := orgIDFromHeader(r, h.defaultOrg)
	existing, err := h.store.GetRule(r.Context(), orgID, ruleID)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	if existing == nil {
		writeAPIError(w, r, http.StatusNotFound, "rule_not_found", "rule not found")
		return
	}

	update := types.Rule{
		RuleID:      ruleID,
		OrgID:       orgID,
		Namespace:   input.Namespace,
		Surface:     input.Surface,
		Name:        input.Name,
		Description: input.Description,
		Action:      input.Action,
		TargetType:  input.TargetType,
		TargetKey:   input.TargetKey,
		ItemIDs:     input.ItemIDs,
		BoostValue:  input.BoostValue,
		MaxPins:     input.MaxPins,
		SegmentID:   input.SegmentID,
		Priority:    input.Priority,
		Enabled:     input.Enabled,
		ValidFrom:   input.ValidFrom,
		ValidUntil:  input.ValidUntil,
	}

	updated, err := h.store.UpdateRule(r.Context(), orgID, update)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	if updated == nil {
		writeAPIError(w, r, http.StatusNotFound, "rule_not_found", "rule not found")
		return
	}

	h.invalidateRuleCache(existing.OrgID, existing.Namespace, existing.Surface)
	h.invalidateRuleCache(updated.OrgID, updated.Namespace, updated.Surface)

	resp := toRuleResponse(*updated)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// RulesList godoc
// @Summary      List merchandising rules
// @Description  List merchandising rules with optional filtering
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        namespace query string true "Namespace"
// @Param        surface query string false "Filter by surface"
// @Param        segment_id query string false "Filter by segment ID"
// @Param        enabled query boolean false "Filter by enabled status"
// @Param        active_now query boolean false "Filter by active status"
// @Param        action query string false "Filter by action type"
// @Param        target_type query string false "Filter by target type"
// @Success      200 {object} types.RulesListResponse
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/admin/rules [get]
func (h *RulesHandler) RulesList(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromHeader(r, h.defaultOrg)
	q := r.URL.Query()
	namespace := strings.TrimSpace(q.Get("namespace"))
	if namespace == "" {
		namespace = "default"
	}

	filters := types.RuleListFilters{}
	if surface := strings.TrimSpace(q.Get("surface")); surface != "" {
		filters.Surface = surface
	}
	if segment := strings.TrimSpace(q.Get("segment_id")); segment != "" {
		if strings.EqualFold(segment, "null") {
			filters.SegmentID = "__NULL__"
		} else {
			filters.SegmentID = segment
		}
	}
	if enabledRaw := strings.TrimSpace(q.Get("enabled")); enabledRaw != "" {
		enabled, err := strconv.ParseBool(enabledRaw)
		if err != nil {
			common.BadRequest(w, r, "invalid_enabled", "enabled must be true or false", nil)
			return
		}
		filters.Enabled = &enabled
	}
	if active := strings.TrimSpace(q.Get("active_now")); strings.EqualFold(active, "true") {
		now := time.Now().UTC()
		filters.ActiveAt = &now
	}
	if action := strings.TrimSpace(q.Get("action")); action != "" {
		act, err := parseRuleAction(action)
		if err != nil {
			common.BadRequest(w, r, "invalid_action", err.Error(), nil)
			return
		}
		filters.Action = &act
	}
	if target := strings.TrimSpace(q.Get("target_type")); target != "" {
		tgt, err := parseRuleTarget(target)
		if err != nil {
			common.BadRequest(w, r, "invalid_target_type", err.Error(), nil)
			return
		}
		filters.TargetType = &tgt
	}

	rulesList, err := h.store.ListRules(r.Context(), orgID, namespace, filters)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	resp := specstypes.RulesListResponse{Rules: make([]specstypes.RuleResponse, 0, len(rulesList))}
	for _, rule := range rulesList {
		resp.Rules = append(resp.Rules, toRuleResponse(rule))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// RulesDryRun godoc
// @Summary      Preview rule effects
// @Description  Preview matched rules and effects without mutating state
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        payload body types.RuleDryRunRequest true "Dry run request"
// @Success      200 {object} types.RuleDryRunResponse
// @Failure      400 {object} common.APIError
// @Failure      500 {object} common.APIError
// @Router       /v1/admin/rules/dry-run [post]
func (h *RulesHandler) RulesDryRun(w http.ResponseWriter, r *http.Request) {
	if h.manager == nil || !h.manager.Enabled() {
		common.BadRequest(w, r, "rules_disabled", "rules engine is disabled", nil)
		return
	}

	var payload specstypes.RuleDryRunRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	payload.Namespace = strings.TrimSpace(payload.Namespace)
	payload.Surface = strings.TrimSpace(payload.Surface)
	payload.SegmentID = strings.TrimSpace(payload.SegmentID)
	if payload.Namespace == "" {
		payload.Namespace = "default"
	}
	if payload.Surface == "" {
		common.BadRequest(w, r, "invalid_surface", "surface is required", nil)
		return
	}

	orgID := orgIDFromHeader(r, h.defaultOrg)
	candidates := make([]recmodel.ScoredItem, 0, len(payload.Items))
	queryIDs := make([]string, 0, len(payload.Items))
	for _, id := range payload.Items {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		candidates = append(candidates, recmodel.ScoredItem{ItemID: trimmed, Score: 0})
		queryIDs = append(queryIDs, trimmed)
	}

	tagMap, err := h.store.ListItemsTags(r.Context(), orgID, payload.Namespace, queryIDs)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	simpleTags := make(map[string][]string, len(tagMap))
	for id, tg := range tagMap {
		simpleTags[id] = append([]string(nil), tg.Tags...)
	}

	evalReq := rules.EvaluateRequest{
		OrgID:               orgID,
		Namespace:           payload.Namespace,
		Surface:             payload.Surface,
		SegmentID:           payload.SegmentID,
		Now:                 time.Now().UTC(),
		Candidates:          candidates,
		ItemTags:            simpleTags,
		BrandTagPrefixes:    h.brandTagPrefixes,
		CategoryTagPrefixes: h.categoryTagPrefixes,
	}

	evalResult, err := h.manager.Evaluate(r.Context(), evalReq)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	resp := specstypes.RuleDryRunResponse{
		RulesEvaluated: make([]string, 0, len(evalResult.EvaluatedRuleIDs)),
		RulesMatched:   make([]specstypes.RuleMatchResponse, 0, len(evalResult.Matches)),
		ItemEffects:    make(map[string]specstypes.RuleItemEffectResponse, len(evalResult.ItemEffects)),
	}
	for _, id := range evalResult.EvaluatedRuleIDs {
		resp.RulesEvaluated = append(resp.RulesEvaluated, id.String())
	}
	for _, match := range evalResult.Matches {
		resp.RulesMatched = append(resp.RulesMatched, specstypes.RuleMatchResponse{
			RuleID:  match.RuleID.String(),
			Action:  string(match.Action),
			Target:  string(match.Target),
			ItemIDs: append([]string(nil), match.ItemIDs...),
		})
	}
	for id, eff := range evalResult.ItemEffects {
		resp.ItemEffects[id] = specstypes.RuleItemEffectResponse{
			Blocked:    eff.Blocked,
			Pinned:     eff.Pinned,
			BoostDelta: eff.BoostDelta,
		}
	}
	if len(evalResult.ReasonTags) > 0 {
		resp.ReasonTags = make(map[string][]string, len(evalResult.ReasonTags))
		for id, tags := range evalResult.ReasonTags {
			resp.ReasonTags[id] = append([]string(nil), tags...)
		}
	}
	if len(evalResult.Pinned) > 0 {
		resp.PinnedPreview = make([]specstypes.RuleDryRunPinnedItem, 0, len(evalResult.Pinned))
		for _, item := range evalResult.Pinned {
			preview := specstypes.RuleDryRunPinnedItem{
				ItemID:         item.ItemID,
				FromCandidates: item.FromCandidates,
			}
			if len(item.Rules) > 0 {
				preview.RuleIDs = make([]string, 0, len(item.Rules))
				for _, rid := range item.Rules {
					preview.RuleIDs = append(preview.RuleIDs, rid.String())
				}
			}
			resp.PinnedPreview = append(resp.PinnedPreview, preview)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *RulesHandler) invalidateRuleCache(orgID uuid.UUID, namespace, surface string) {
	if h.manager == nil {
		return
	}
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		namespace = "default"
	}
	h.manager.Invalidate(orgID, namespace, strings.TrimSpace(surface))
}

func writeAPIError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	ae := common.NewAPIError(code, message, status)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ae)
}

// Helper types and functions copied from legacy handler with minor adjustments.
type ruleInput struct {
	Namespace   string
	Surface     string
	Name        string
	Description string
	Action      types.RuleAction
	TargetType  types.RuleTarget
	TargetKey   string
	ItemIDs     []string
	BoostValue  *float64
	MaxPins     *int
	SegmentID   string
	Priority    int
	Enabled     bool
	ValidFrom   *time.Time
	ValidUntil  *time.Time
}

func parseRulePayload(payload specstypes.RulePayload) (ruleInput, error) {
	input := ruleInput{}
	input.Namespace = strings.TrimSpace(payload.Namespace)
	if input.Namespace == "" {
		input.Namespace = "default"
	}
	input.Surface = strings.TrimSpace(payload.Surface)
	if input.Surface == "" {
		return ruleInput{}, errors.New("surface is required")
	}
	input.Name = strings.TrimSpace(payload.Name)
	if input.Name == "" {
		return ruleInput{}, errors.New("name is required")
	}
	input.Description = strings.TrimSpace(payload.Description)

	action, err := parseRuleAction(payload.Action)
	if err != nil {
		return ruleInput{}, err
	}
	targetType, err := parseRuleTarget(payload.TargetType)
	if err != nil {
		return ruleInput{}, err
	}
	input.Action = action
	input.TargetType = targetType

	targetKey := strings.TrimSpace(payload.TargetKey)
	if targetType != types.RuleTargetItem && targetKey == "" {
		return ruleInput{}, errors.New("target_key is required for non-item targets")
	}
	input.TargetKey = targetKey

	itemIDs := make([]string, 0, len(payload.ItemIDs))
	seen := make(map[string]struct{}, len(payload.ItemIDs))
	for _, id := range payload.ItemIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		itemIDs = append(itemIDs, trimmed)
	}
	if targetType == types.RuleTargetItem && len(itemIDs) == 0 {
		return ruleInput{}, errors.New("item_ids must be provided for item target")
	}
	input.ItemIDs = itemIDs

	if action == types.RuleActionBoost {
		if payload.BoostValue == nil || *payload.BoostValue == 0 {
			return ruleInput{}, errors.New("boost_value must be non-zero for BOOST action")
		}
		input.BoostValue = payload.BoostValue
	} else if payload.BoostValue != nil {
		return ruleInput{}, errors.New("boost_value is only valid for BOOST action")
	}

	if action == types.RuleActionPin {
		if len(itemIDs) == 0 && targetType != types.RuleTargetItem {
			return ruleInput{}, errors.New("item_ids required for PIN action")
		}
		if payload.MaxPins != nil {
			if *payload.MaxPins <= 0 {
				return ruleInput{}, errors.New("max_pins must be positive when provided")
			}
			input.MaxPins = payload.MaxPins
		}
	} else if payload.MaxPins != nil {
		return ruleInput{}, errors.New("max_pins is only valid for PIN action")
	}

	if payload.Priority != nil {
		input.Priority = *payload.Priority
	}
	if payload.Enabled != nil {
		input.Enabled = *payload.Enabled
	} else {
		input.Enabled = true
	}
	input.SegmentID = strings.TrimSpace(payload.SegmentID)

	if payload.ValidFrom != "" {
		ts, err := time.Parse(time.RFC3339, strings.TrimSpace(payload.ValidFrom))
		if err != nil {
			return ruleInput{}, errors.New("valid_from must be RFC3339 timestamp")
		}
		input.ValidFrom = &ts
	}
	if payload.ValidUntil != "" {
		ts, err := time.Parse(time.RFC3339, strings.TrimSpace(payload.ValidUntil))
		if err != nil {
			return ruleInput{}, errors.New("valid_until must be RFC3339 timestamp")
		}
		input.ValidUntil = &ts
	}
	if input.ValidFrom != nil && input.ValidUntil != nil && !input.ValidUntil.After(*input.ValidFrom) {
		return ruleInput{}, errors.New("valid_until must be after valid_from")
	}

	return input, nil
}

func parseRuleAction(raw string) (types.RuleAction, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case string(types.RuleActionBlock):
		return types.RuleActionBlock, nil
	case string(types.RuleActionPin):
		return types.RuleActionPin, nil
	case string(types.RuleActionBoost):
		return types.RuleActionBoost, nil
	default:
		return "", errors.New("action must be one of BLOCK, PIN, BOOST")
	}
}

func parseRuleTarget(raw string) (types.RuleTarget, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case string(types.RuleTargetItem):
		return types.RuleTargetItem, nil
	case string(types.RuleTargetTag):
		return types.RuleTargetTag, nil
	case string(types.RuleTargetBrand):
		return types.RuleTargetBrand, nil
	case string(types.RuleTargetCategory):
		return types.RuleTargetCategory, nil
	default:
		return "", errors.New("target_type must be one of ITEM, TAG, BRAND, CATEGORY")
	}
}

func toRuleResponse(rule types.Rule) specstypes.RuleResponse {
	resp := specstypes.RuleResponse{
		RuleID:      rule.RuleID.String(),
		Namespace:   rule.Namespace,
		Surface:     rule.Surface,
		Name:        rule.Name,
		Description: rule.Description,
		Action:      string(rule.Action),
		TargetType:  string(rule.TargetType),
		TargetKey:   rule.TargetKey,
		ItemIDs:     append([]string(nil), rule.ItemIDs...),
		BoostValue:  rule.BoostValue,
		MaxPins:     rule.MaxPins,
		SegmentID:   rule.SegmentID,
		Priority:    rule.Priority,
		Enabled:     rule.Enabled,
		ValidFrom:   formatTime(rule.ValidFrom),
		ValidUntil:  formatTime(rule.ValidUntil),
		CreatedAt:   rule.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   rule.UpdatedAt.UTC().Format(time.RFC3339),
	}
	return resp
}

func formatTime(ts *time.Time) string {
	if ts == nil {
		return ""
	}
	return ts.UTC().Format(time.RFC3339)
}
