package handlers

import (
	"strings"
	"time"

	"recsys/internal/algorithm"
)

// RecommendationConfig captures the base algorithm tunables used for ranking.
type RecommendationConfig struct {
	HalfLifeDays                  float64
	CoVisWindowDays               float64
	PopularityFanout              int
	MMRLambda                     float64
	BrandCap                      int
	CategoryCap                   int
	RuleExcludeEvents             bool
	ExcludeEventTypes             []int16
	BrandTagPrefixes              []string
	CategoryTagPrefixes           []string
	PurchasedWindowDays           float64
	ProfileWindowDays             float64
	ProfileBoost                  float64
	ProfileTopNTags               int
	ProfileMinEventsForBoost      int
	ProfileColdStartMultiplier    float64
	ProfileStarterBlendWeight     float64
	MMRPresets                    map[string]float64
	BlendAlpha                    float64
	BlendBeta                     float64
	BlendGamma                    float64
	NewUserBlendAlpha             *float64
	NewUserBlendBeta              *float64
	NewUserBlendGamma             *float64
	NewUserMMRLambda              *float64
	NewUserPopFanout              *int
	BanditExperiment              BanditExperimentConfig
	RulesEnabled                  bool
	CoverageCacheTTL              time.Duration
	CoverageLongTailHintThreshold float64
}

// Clone returns a deep copy of the config to avoid sharing slice/map references.
func (c RecommendationConfig) Clone() RecommendationConfig {
	clone := c
	if len(c.ExcludeEventTypes) > 0 {
		clone.ExcludeEventTypes = append([]int16(nil), c.ExcludeEventTypes...)
	}
	if len(c.BrandTagPrefixes) > 0 {
		clone.BrandTagPrefixes = append([]string(nil), c.BrandTagPrefixes...)
	}
	if len(c.CategoryTagPrefixes) > 0 {
		clone.CategoryTagPrefixes = append([]string(nil), c.CategoryTagPrefixes...)
	}
	if len(c.MMRPresets) > 0 {
		clone.MMRPresets = make(map[string]float64, len(c.MMRPresets))
		for k, v := range c.MMRPresets {
			clone.MMRPresets[k] = v
		}
	}
	if c.NewUserBlendAlpha != nil {
		val := *c.NewUserBlendAlpha
		clone.NewUserBlendAlpha = &val
	}
	if c.NewUserBlendBeta != nil {
		val := *c.NewUserBlendBeta
		clone.NewUserBlendBeta = &val
	}
	if c.NewUserBlendGamma != nil {
		val := *c.NewUserBlendGamma
		clone.NewUserBlendGamma = &val
	}
	if c.NewUserMMRLambda != nil {
		val := *c.NewUserMMRLambda
		clone.NewUserMMRLambda = &val
	}
	if c.NewUserPopFanout != nil {
		val := *c.NewUserPopFanout
		clone.NewUserPopFanout = &val
	}
	if len(c.BanditExperiment.Surfaces) > 0 {
		clone.BanditExperiment.Surfaces = make(map[string]struct{}, len(c.BanditExperiment.Surfaces))
		for k := range c.BanditExperiment.Surfaces {
			clone.BanditExperiment.Surfaces[k] = struct{}{}
		}
	} else if c.BanditExperiment.Surfaces != nil {
		clone.BanditExperiment.Surfaces = nil
	}
	return clone
}

// BanditExperimentConfig controls exploration holdouts for experiments.
type BanditExperimentConfig struct {
	Enabled        bool
	HoldoutPercent float64
	Surfaces       map[string]struct{}
	Label          string
}

// Applies returns true if the experiment applies to the given surface.
func (c BanditExperimentConfig) Applies(surface string) bool {
	if !c.Enabled {
		return false
	}
	if len(c.Surfaces) == 0 {
		return true
	}
	normalized := strings.ToLower(strings.TrimSpace(surface))
	if normalized == "" {
		normalized = "default"
	}
	_, ok := c.Surfaces[normalized]
	return ok
}

// BaseConfig materializes the config into an algorithm.Config with defensive copies.
func (c RecommendationConfig) BaseConfig() algorithm.Config {
	cfg := algorithm.Config{
		BlendAlpha:                 c.BlendAlpha,
		BlendBeta:                  c.BlendBeta,
		BlendGamma:                 c.BlendGamma,
		ProfileBoost:               c.ProfileBoost,
		ProfileWindowDays:          c.ProfileWindowDays,
		ProfileTopNTags:            c.ProfileTopNTags,
		ProfileMinEventsForBoost:   c.ProfileMinEventsForBoost,
		ProfileColdStartMultiplier: c.ProfileColdStartMultiplier,
		ProfileStarterBlendWeight:  c.ProfileStarterBlendWeight,
		MMRLambda:                  c.MMRLambda,
		BrandCap:                   c.BrandCap,
		CategoryCap:                c.CategoryCap,
		HalfLifeDays:               c.HalfLifeDays,
		CoVisWindowDays:            int(c.CoVisWindowDays),
		PurchasedWindowDays:        int(c.PurchasedWindowDays),
		RuleExcludeEvents:          c.RuleExcludeEvents,
		PopularityFanout:           c.PopularityFanout,
		RulesEnabled:               c.RulesEnabled,
	}

	if len(c.ExcludeEventTypes) > 0 {
		cfg.ExcludeEventTypes = append([]int16(nil), c.ExcludeEventTypes...)
	}
	if len(c.BrandTagPrefixes) > 0 {
		cfg.BrandTagPrefixes = append([]string(nil), c.BrandTagPrefixes...)
	}
	if len(c.CategoryTagPrefixes) > 0 {
		cfg.CategoryTagPrefixes = append([]string(nil), c.CategoryTagPrefixes...)
	}
	return cfg
}
