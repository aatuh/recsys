package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"recsys/internal/http/common"
	handlerstypes "recsys/specs/types"

	"go.uber.org/zap"
)

// RecommendationConfigAdminHandler exposes GET/POST endpoints for config management.
type RecommendationConfigAdminHandler struct {
	manager   *RecommendationConfigManager
	logger    *zap.Logger
	listeners []func(RecommendationConfig)
}

// NewRecommendationAdminConfigHandler constructs the handler.
func NewRecommendationAdminConfigHandler(manager *RecommendationConfigManager, logger *zap.Logger) *RecommendationConfigAdminHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RecommendationConfigAdminHandler{
		manager: manager,
		logger:  logger,
	}
}

// RegisterListener registers a callback invoked after successful updates.
func (h *RecommendationConfigAdminHandler) RegisterListener(fn func(RecommendationConfig)) {
	if fn == nil {
		return
	}
	h.listeners = append(h.listeners, fn)
}

// RecommendationConfigGet godoc
// @Summary      Fetch the active recommendation config snapshot
// @Tags         admin
// @Produce      json
// @Param        namespace query string false "Namespace" default(default)
// @Success      200 {object} types.RecommendationConfigDocument
// @Router       /v1/admin/recommendation/config [get]
func (h *RecommendationConfigAdminHandler) Get(w http.ResponseWriter, r *http.Request) {
	snap := h.manager.Snapshot()
	resp := handlerstypes.RecommendationConfigDocument{
		Namespace: "default",
		Config:    convertConfigToSpec(snap.Config),
		Metadata: handlerstypes.RecommendationConfigMetadata{
			UpdatedAt: snap.Metadata.UpdatedAt,
			UpdatedBy: snap.Metadata.UpdatedBy,
			Source:    snap.Metadata.Source,
			Notes:     snap.Metadata.Notes,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// RecommendationConfigUpdate godoc
// @Summary      Update the active recommendation config
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        payload body types.RecommendationConfigUpdateRequest true "Updated config"
// @Success      200 {object} types.RecommendationConfigDocument
// @Router       /v1/admin/recommendation/config [post]
func (h *RecommendationConfigAdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req handlerstypes.RecommendationConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpErrorWithLogger(w, r, err, http.StatusBadRequest, h.logger)
		return
	}
	if req.Config == nil {
		common.BadRequest(w, r, "missing_config", "config payload is required", nil)
		return
	}

	nextCfg, err := convertSpecToConfig(req.Config)
	if err != nil {
		common.BadRequest(w, r, "invalid_config", err.Error(), nil)
		return
	}

	meta := RecommendationConfigMetadata{
		UpdatedBy: req.Author,
		Notes:     req.Notes,
		Source:    "api",
	}
	snap := h.manager.Update(nextCfg, meta)
	h.notifyListeners(snap.Config)
	resp := handlerstypes.RecommendationConfigDocument{
		Namespace: "default",
		Config:    convertConfigToSpec(snap.Config),
		Metadata: handlerstypes.RecommendationConfigMetadata{
			UpdatedAt: snap.Metadata.UpdatedAt,
			UpdatedBy: snap.Metadata.UpdatedBy,
			Source:    snap.Metadata.Source,
			Notes:     snap.Metadata.Notes,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *RecommendationConfigAdminHandler) notifyListeners(cfg RecommendationConfig) {
	for _, fn := range h.listeners {
		fn(cfg.Clone())
	}
}

func convertConfigToSpec(cfg RecommendationConfig) *handlerstypes.RecommendationConfigPayload {
	payload := &handlerstypes.RecommendationConfigPayload{
		HalfLifeDays:               cfg.HalfLifeDays,
		CoVisWindowDays:            cfg.CoVisWindowDays,
		PopularityFanout:           cfg.PopularityFanout,
		MaxK:                       cfg.MaxK,
		MaxFanout:                  cfg.MaxFanout,
		MaxExcludeIDs:              cfg.MaxExcludeIDs,
		MaxAnchorsInjected:         cfg.MaxAnchorsInjected,
		MMRLambda:                  cfg.MMRLambda,
		BrandCap:                   cfg.BrandCap,
		CategoryCap:                cfg.CategoryCap,
		RuleExcludeEvents:          cfg.RuleExcludeEvents,
		ExcludeEventTypes:          cfg.ExcludeEventTypes,
		BrandTagPrefixes:           cfg.BrandTagPrefixes,
		CategoryTagPrefixes:        cfg.CategoryTagPrefixes,
		PurchasedWindowDays:        cfg.PurchasedWindowDays,
		ProfileWindowDays:          cfg.ProfileWindowDays,
		ProfileBoost:               cfg.ProfileBoost,
		ProfileTopNTags:            cfg.ProfileTopNTags,
		ProfileMinEventsForBoost:   cfg.ProfileMinEventsForBoost,
		ProfileColdStartMultiplier: cfg.ProfileColdStartMultiplier,
		ProfileStarterBlendWeight:  cfg.ProfileStarterBlendWeight,
		MMRPresets:                 cfg.MMRPresets,
		BlendAlpha:                 cfg.BlendAlpha,
		BlendBeta:                  cfg.BlendBeta,
		BlendGamma:                 cfg.BlendGamma,
		NewUserMMRLambda:           cfg.NewUserMMRLambda,
		NewUserBlendAlpha:          cfg.NewUserBlendAlpha,
		NewUserBlendBeta:           cfg.NewUserBlendBeta,
		NewUserBlendGamma:          cfg.NewUserBlendGamma,
		NewUserPopFanout:           cfg.NewUserPopFanout,
		BanditExperiment: handlerstypes.BanditExperimentConfig{
			Enabled:        cfg.BanditExperiment.Enabled,
			HoldoutPercent: cfg.BanditExperiment.HoldoutPercent,
			Label:          cfg.BanditExperiment.Label,
			Surfaces:       mapKeys(cfg.BanditExperiment.Surfaces),
		},
		RulesEnabled:                  cfg.RulesEnabled,
		CoverageCacheTTLSeconds:       cfg.CoverageCacheTTL.Seconds(),
		CoverageLongTailHintThreshold: cfg.CoverageLongTailHintThreshold,
	}
	if len(cfg.SegmentProfiles) > 0 {
		payload.SegmentProfiles = make(map[string]handlerstypes.SegmentProfileConfigPayload, len(cfg.SegmentProfiles))
		for segment, profile := range cfg.SegmentProfiles {
			payload.SegmentProfiles[segment] = convertSegmentProfileToSpec(profile)
		}
	}
	return payload
}

func convertSpecToConfig(payload *handlerstypes.RecommendationConfigPayload) (RecommendationConfig, error) {
	cfg := RecommendationConfig{
		HalfLifeDays:                  payload.HalfLifeDays,
		CoVisWindowDays:               payload.CoVisWindowDays,
		PopularityFanout:              payload.PopularityFanout,
		MaxK:                          payload.MaxK,
		MaxFanout:                     payload.MaxFanout,
		MaxExcludeIDs:                 payload.MaxExcludeIDs,
		MaxAnchorsInjected:            payload.MaxAnchorsInjected,
		MMRLambda:                     payload.MMRLambda,
		BrandCap:                      payload.BrandCap,
		CategoryCap:                   payload.CategoryCap,
		RuleExcludeEvents:             payload.RuleExcludeEvents,
		ExcludeEventTypes:             payload.ExcludeEventTypes,
		BrandTagPrefixes:              payload.BrandTagPrefixes,
		CategoryTagPrefixes:           payload.CategoryTagPrefixes,
		PurchasedWindowDays:           payload.PurchasedWindowDays,
		ProfileWindowDays:             payload.ProfileWindowDays,
		ProfileBoost:                  payload.ProfileBoost,
		ProfileTopNTags:               payload.ProfileTopNTags,
		ProfileMinEventsForBoost:      payload.ProfileMinEventsForBoost,
		ProfileColdStartMultiplier:    payload.ProfileColdStartMultiplier,
		ProfileStarterBlendWeight:     payload.ProfileStarterBlendWeight,
		MMRPresets:                    payload.MMRPresets,
		BlendAlpha:                    payload.BlendAlpha,
		BlendBeta:                     payload.BlendBeta,
		BlendGamma:                    payload.BlendGamma,
		RulesEnabled:                  payload.RulesEnabled,
		CoverageCacheTTL:              time.Duration(payload.CoverageCacheTTLSeconds * float64(time.Second)),
		CoverageLongTailHintThreshold: payload.CoverageLongTailHintThreshold,
		BanditExperiment: BanditExperimentConfig{
			Enabled:        payload.BanditExperiment.Enabled,
			HoldoutPercent: payload.BanditExperiment.HoldoutPercent,
			Label:          payload.BanditExperiment.Label,
		},
	}
	if payload.NewUserBlendAlpha != nil {
		val := *payload.NewUserBlendAlpha
		cfg.NewUserBlendAlpha = &val
	}
	if payload.NewUserBlendBeta != nil {
		val := *payload.NewUserBlendBeta
		cfg.NewUserBlendBeta = &val
	}
	if payload.NewUserBlendGamma != nil {
		val := *payload.NewUserBlendGamma
		cfg.NewUserBlendGamma = &val
	}
	if payload.NewUserMMRLambda != nil {
		val := *payload.NewUserMMRLambda
		cfg.NewUserMMRLambda = &val
	}
	if payload.NewUserPopFanout != nil {
		val := *payload.NewUserPopFanout
		cfg.NewUserPopFanout = &val
	}
	if len(payload.BanditExperiment.Surfaces) > 0 {
		cfg.BanditExperiment.Surfaces = make(map[string]struct{}, len(payload.BanditExperiment.Surfaces))
		for _, v := range payload.BanditExperiment.Surfaces {
			cfg.BanditExperiment.Surfaces[strings.ToLower(strings.TrimSpace(v))] = struct{}{}
		}
	}
	if len(payload.SegmentProfiles) > 0 {
		cfg.SegmentProfiles = make(map[string]SegmentProfileConfig, len(payload.SegmentProfiles))
		for segment, profile := range payload.SegmentProfiles {
			key := strings.ToLower(strings.TrimSpace(segment))
			if key == "" {
				continue
			}
			cfg.SegmentProfiles[key] = convertSpecToSegmentProfile(profile)
		}
	}
	return cfg, nil
}

func convertSegmentProfileToSpec(profile SegmentProfileConfig) handlerstypes.SegmentProfileConfigPayload {
	payload := handlerstypes.SegmentProfileConfigPayload{
		BlendAlpha: profile.BlendAlpha,
		BlendBeta:  profile.BlendBeta,
		BlendGamma: profile.BlendGamma,
	}
	if profile.MMRLambda != nil {
		val := *profile.MMRLambda
		payload.MMRLambda = &val
	}
	if profile.PopularityFanout != nil {
		val := *profile.PopularityFanout
		payload.PopularityFanout = &val
	}
	if profile.ProfileStarterBlendWeight != nil {
		val := *profile.ProfileStarterBlendWeight
		payload.ProfileStarterBlendWeight = &val
	}
	return payload
}

func convertSpecToSegmentProfile(payload handlerstypes.SegmentProfileConfigPayload) SegmentProfileConfig {
	profile := SegmentProfileConfig{
		BlendAlpha: payload.BlendAlpha,
		BlendBeta:  payload.BlendBeta,
		BlendGamma: payload.BlendGamma,
	}
	if payload.MMRLambda != nil {
		val := *payload.MMRLambda
		profile.MMRLambda = &val
	}
	if payload.PopularityFanout != nil {
		val := *payload.PopularityFanout
		profile.PopularityFanout = &val
	}
	if payload.ProfileStarterBlendWeight != nil {
		val := *payload.ProfileStarterBlendWeight
		profile.ProfileStarterBlendWeight = &val
	}
	return profile
}

func mapKeys(input map[string]struct{}) []string {
	if len(input) == 0 {
		return nil
	}
	out := make([]string, 0, len(input))
	for k := range input {
		out = append(out, k)
	}
	return out
}
