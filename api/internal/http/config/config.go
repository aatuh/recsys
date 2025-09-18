package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"recsys/internal/types"
	"recsys/shared/util"

	"github.com/google/uuid"
)

type Config struct {
	DatabaseURL         string
	DefaultOrgID        uuid.UUID
	HalfLifeDays        float64 // popularity decay half-life
	CoVisWindowDays     float64 // default co-vis window (e.g., 30)
	PopularityFanout    int     // optional prefilter cap for popularity
	MMRLambda           float64 // 0..1; 0 disables MMR
	BrandCap            int     // max items per brand:* tag; 0 disables
	CategoryCap         int     // max items per category:* or cat:* tag; 0 disables
	RuleExcludeEvents   bool    // if true, exclude user's recent events
	ExcludeEventTypes   []int16 // event types excluded from recommendations when rule enabled
	BrandTagPrefixes    []string
	CategoryTagPrefixes []string
	PurchasedWindowDays float64 // lookback window for purchases (days)
	ProfileWindowDays   float64 // lookback for building profile; <=0 disables windowing
	ProfileBoost        float64 // multiplier in [0, +inf). 0 disables personalization
	ProfileTopNTags     int     // limit of profile tags considered
	BlendAlpha          float64
	BlendBeta           float64
	BlendGamma          float64
	BanditAlgo          types.Algorithm

	DecisionTraceEnabled          bool
	DecisionTraceQueueSize        int
	DecisionTraceBatchSize        int
	DecisionTraceFlushInterval    time.Duration
	DecisionTraceSampleDefault    float64
	DecisionTraceNamespaceSamples map[string]float64
	DecisionTraceSalt             string
}

func Load() (Config, error) {
	var c Config
	c.DatabaseURL = util.MustGetEnv("DATABASE_URL")
	parsePrefixes := func(raw string) []string {
		parts := strings.Split(raw, ",")
		out := make([]string, 0, len(parts))
		seen := make(map[string]struct{})
		for _, p := range parts {
			trimmed := strings.ToLower(strings.TrimSpace(p))
			trimmed = strings.TrimSuffix(trimmed, ":")
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			out = append(out, trimmed)
		}
		return out
	}

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

	if prefixes := parsePrefixes(util.MustGetEnv("BRAND_TAG_PREFIXES")); len(prefixes) > 0 {
		c.BrandTagPrefixes = prefixes
	} else {
		return c, errors.New("BRAND_TAG_PREFIXES must include at least one prefix")
	}
	if prefixes := parsePrefixes(util.MustGetEnv("CATEGORY_TAG_PREFIXES")); len(prefixes) > 0 {
		c.CategoryTagPrefixes = prefixes
	} else {
		return c, errors.New("CATEGORY_TAG_PREFIXES must include at least one prefix")
	}

	// Business rule: exclude purchased items in a window.
	c.RuleExcludeEvents = util.MustGetEnv("RULE_EXCLUDE_EVENTS") == "true"
	rawTypes := util.MustGetEnv("EXCLUDE_EVENT_TYPES")
	parts := strings.Split(rawTypes, ",")
	c.ExcludeEventTypes = make([]int16, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		iv, err := strconv.Atoi(p)
		if err != nil || iv < -32768 || iv > 32767 {
			return c, errors.New("EXCLUDE_EVENT_TYPES must be a comma-separated list of valid int16 values")
		}
		c.ExcludeEventTypes = append(c.ExcludeEventTypes, int16(iv))
	}
	if len(c.ExcludeEventTypes) == 0 {
		c.ExcludeEventTypes = nil
	}

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
		c.BanditAlgo = types.AlgorithmThompson
	case "ucb1":
		c.BanditAlgo = types.AlgorithmUCB1
	default:
		return c, errors.New("BANDIT_ALGO must be 'thompson' or 'ucb1'")
	}

	c.DecisionTraceEnabled = util.MustGetEnv("AUDIT_DECISIONS_ENABLED") == "true"
	if c.DecisionTraceEnabled {
		defaultRate, err := strconv.ParseFloat(util.MustGetEnv("AUDIT_DECISIONS_SAMPLE_DEFAULT"), 64)
		if err != nil || defaultRate < 0 || defaultRate > 1 {
			return c, errors.New("AUDIT_DECISIONS_SAMPLE_DEFAULT must be between 0 and 1")
		}
		c.DecisionTraceSampleDefault = defaultRate

		queueSize, err := strconv.Atoi(util.MustGetEnv("AUDIT_DECISIONS_QUEUE"))
		if err != nil || queueSize <= 0 {
			return c, errors.New("AUDIT_DECISIONS_QUEUE must be a positive integer")
		}
		c.DecisionTraceQueueSize = queueSize

		batchSize, err := strconv.Atoi(util.MustGetEnv("AUDIT_DECISIONS_BATCH"))
		if err != nil || batchSize <= 0 {
			return c, errors.New("AUDIT_DECISIONS_BATCH must be a positive integer")
		}
		c.DecisionTraceBatchSize = batchSize

		flushStr := util.MustGetEnv("AUDIT_DECISIONS_FLUSH_INTERVAL")
		flushDur, err := time.ParseDuration(strings.TrimSpace(flushStr))
		if err != nil || flushDur <= 0 {
			return c, errors.New("AUDIT_DECISIONS_FLUSH_INTERVAL must be a positive duration (e.g. 250ms)")
		}
		c.DecisionTraceFlushInterval = flushDur

		rawOverrides := strings.TrimSpace(util.MustGetEnv("AUDIT_DECISIONS_SAMPLE_OVERRIDES"))
		if rawOverrides != "-" {
			overrides := make(map[string]float64)
			parts := strings.Split(rawOverrides, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				kv := strings.SplitN(part, "=", 2)
				if len(kv) != 2 {
					return c, fmt.Errorf("AUDIT_DECISIONS_SAMPLE_OVERRIDES entry %q must be namespace=rate", part)
				}
				ns := strings.TrimSpace(kv[0])
				if ns == "" {
					return c, errors.New("AUDIT_DECISIONS_SAMPLE_OVERRIDES namespace cannot be empty")
				}
				val, err := strconv.ParseFloat(strings.TrimSpace(kv[1]), 64)
				if err != nil || val < 0 || val > 1 {
					return c, fmt.Errorf("AUDIT_DECISIONS_SAMPLE_OVERRIDES invalid rate for namespace %q", ns)
				}
				overrides[ns] = val
			}
			if len(overrides) > 0 {
				c.DecisionTraceNamespaceSamples = overrides
			}
		}

		salt := strings.TrimSpace(util.MustGetEnv("AUDIT_DECISIONS_SALT"))
		if salt == "" {
			return c, errors.New("AUDIT_DECISIONS_SALT must be non-empty")
		}
		c.DecisionTraceSalt = salt
	}

	return c, nil
}
