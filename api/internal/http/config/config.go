package config

import (
	"errors"
	"strconv"
	"strings"

	"recsys/internal/bandit"
	"recsys/shared/util"

	"github.com/google/uuid"
)

type Config struct {
	DatabaseURL          string
	DefaultOrgID         uuid.UUID
	HalfLifeDays         float64 // popularity decay half-life
	CoVisWindowDays      float64 // default co-vis window (e.g., 30)
	PopularityFanout     int     // optional prefilter cap for popularity
	MMRLambda            float64 // 0..1; 0 disables MMR
	BrandCap             int     // max items per brand:* tag; 0 disables
	CategoryCap          int     // max items per category:* or cat:* tag; 0 disables
	RuleExcludePurchased bool    // if true, exclude user's purchased items
	PurchasedWindowDays  float64 // lookback window for purchases (days)
	ProfileWindowDays    float64 // lookback for building profile; <=0 disables windowing
	ProfileBoost         float64 // multiplier in [0, +inf). 0 disables personalization
	ProfileTopNTags      int     // limit of profile tags considered
	BlendAlpha           float64
	BlendBeta            float64
	BlendGamma           float64
	BanditAlgo           bandit.Algorithm
}

func Load() (Config, error) {
	var c Config
	c.DatabaseURL = util.MustGetEnv("DATABASE_URL")

	org := util.MustGetEnv("ORG_ID")
	id, err := uuid.Parse(org)
	if err != nil {
		return c, err
	}
	c.DefaultOrgID = id
	hl := util.MustGetEnv("POPULARITY_HALFLIFE_DAYS")
	v, err := strconv.ParseFloat(hl, 64)
	if err != nil || v <= 0 {
		return c, errors.New("POPULARITY_HALFLIFE_DAYS must be a positive number")
	}
	c.HalfLifeDays = v

	// Optional/tenant-configurable windows; provide sensible defaults.
	if vv, err := strconv.ParseFloat(util.MustGetEnv("COVIS_WINDOW_DAYS"), 64); err == nil && vv > 0 {
		c.CoVisWindowDays = vv
	} else {
		return c, errors.New("COVIS_WINDOW_DAYS must be a positive number")
	}

	// Optional fan-out cap (not strictly required in Stage 1).
	if iv, err := strconv.Atoi(util.MustGetEnv("POPULARITY_FANOUT")); err == nil && iv > 0 {
		c.PopularityFanout = iv
	} else {
		return c, errors.New("POPULARITY_FANOUT must be a positive number")
	}

	// MMR lambda in [0,1]. 0 disables MMR.
	fv, err := strconv.ParseFloat(util.MustGetEnv("MMR_LAMBDA"), 64)
	if err != nil || fv < 0 || fv > 1 {
		return c, errors.New("MMR_LAMBDA must be a float in [0,1]")
	}
	c.MMRLambda = fv

	// Caps. 0 disables.
	iv, err := strconv.Atoi(util.MustGetEnv("BRAND_CAP"))
	if err != nil || iv < 0 {
		return c, errors.New("BRAND_CAP must be a non-negative integer")
	}
	c.BrandCap = iv

	iv, err = strconv.Atoi(util.MustGetEnv("CATEGORY_CAP"))
	if err != nil || iv < 0 {
		return c, errors.New("CATEGORY_CAP must be a non-negative integer")
	}
	c.CategoryCap = iv

	// Business rule: exclude purchased items in a window.
	c.RuleExcludePurchased = util.MustGetEnv("RULE_EXCLUDE_PURCHASED") == "true"

	fv, err = strconv.ParseFloat(util.MustGetEnv("PURCHASED_WINDOW_DAYS"), 64)
	if err != nil || fv <= 0 {
		return c, errors.New("PURCHASED_WINDOW_DAYS must be a positive number")
	}
	c.PurchasedWindowDays = fv

	s := util.MustGetEnv("PROFILE_WINDOW_DAYS")
	f, err := strconv.ParseFloat(s, 64)
	if err == nil && f > 0 {
		c.ProfileWindowDays = f
	} else {
		return c, errors.New("PROFILE_WINDOW_DAYS must be a positive number when set")
	}

	// Boost factor. 0 disables personalization.
	s = util.MustGetEnv("PROFILE_BOOST")
	f, err = strconv.ParseFloat(s, 64)
	if err == nil && f >= 0 {
		c.ProfileBoost = f
	} else {
		return c, errors.New("PROFILE_BOOST must be >= 0")
	}

	// Consider top-N strongest profile tags.
	s = util.MustGetEnv("PROFILE_TOP_N")
	i, err := strconv.Atoi(s)
	if err == nil && i > 0 {
		c.ProfileTopNTags = i
	} else {
		return c, errors.New("PROFILE_TOP_N must be a positive integer")
	}

	// Blended scoring weights (α, β, γ >= 0)
	s = util.MustGetEnv("BLEND_ALPHA")
	if f, err := strconv.ParseFloat(s, 64); err == nil && f >= 0 {
		c.BlendAlpha = f
	} else {
		return c, errors.New("BLEND_ALPHA must be >= 0")
	}
	s = util.MustGetEnv("BLEND_BETA")
	if f, err := strconv.ParseFloat(s, 64); err == nil && f >= 0 {
		c.BlendBeta = f
	} else {
		return c, errors.New("BLEND_BETA must be >= 0")
	}
	s = util.MustGetEnv("BLEND_GAMMA")
	if f, err := strconv.ParseFloat(s, 64); err == nil && f >= 0 {
		c.BlendGamma = f
	} else {
		return c, errors.New("BLEND_GAMMA must be >= 0")
	}

	// Bandit algorithm
	ba := strings.ToLower(util.MustGetEnv("BANDIT_ALGO"))
	switch ba {
	case "thompson":
		c.BanditAlgo = bandit.AlgorithmThompson
	case "ucb1":
		c.BanditAlgo = bandit.AlgorithmUCB1
	default:
		return c, errors.New("BANDIT_ALGO must be 'thompson' or 'ucb1'")
	}

	return c, nil
}
