package handlers

import (
	"strings"

	"recsys/internal/algorithm"
)

// RecommendationConfig captures the base algorithm tunables used for ranking.
type RecommendationConfig struct {
	HalfLifeDays        float64
	CoVisWindowDays     float64
	PopularityFanout    int
	MMRLambda           float64
	BrandCap            int
	CategoryCap         int
	RuleExcludeEvents   bool
	ExcludeEventTypes   []int16
	BrandTagPrefixes    []string
	CategoryTagPrefixes []string
	PurchasedWindowDays float64
	ProfileWindowDays   float64
	ProfileBoost        float64
	ProfileTopNTags     int
	BlendAlpha          float64
	BlendBeta           float64
	BlendGamma          float64
	BanditExperiment    BanditExperimentConfig
	RulesEnabled        bool
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
		BlendAlpha:          c.BlendAlpha,
		BlendBeta:           c.BlendBeta,
		BlendGamma:          c.BlendGamma,
		ProfileBoost:        c.ProfileBoost,
		ProfileWindowDays:   c.ProfileWindowDays,
		ProfileTopNTags:     c.ProfileTopNTags,
		MMRLambda:           c.MMRLambda,
		BrandCap:            c.BrandCap,
		CategoryCap:         c.CategoryCap,
		HalfLifeDays:        c.HalfLifeDays,
		CoVisWindowDays:     int(c.CoVisWindowDays),
		PurchasedWindowDays: int(c.PurchasedWindowDays),
		RuleExcludeEvents:   c.RuleExcludeEvents,
		PopularityFanout:    c.PopularityFanout,
		RulesEnabled:        c.RulesEnabled,
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
