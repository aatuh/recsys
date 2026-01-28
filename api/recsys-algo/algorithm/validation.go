package algorithm

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ValidationError captures a single field validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	if e.Message == "" {
		return e.Field
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors aggregates multiple validation failures.
type ValidationErrors []ValidationError

func (errs ValidationErrors) Error() string {
	if len(errs) == 0 {
		return ""
	}
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Fields returns a copy of the collected validation errors.
func (errs ValidationErrors) Fields() []ValidationError {
	if len(errs) == 0 {
		return nil
	}
	out := make([]ValidationError, len(errs))
	copy(out, errs)
	return out
}

// Validate checks configuration invariants.
func (c Config) Validate() error {
	var errs ValidationErrors
	if c.BlendAlpha < 0 {
		errs = append(errs, ValidationError{Field: "blend_alpha", Message: "must be >= 0"})
	}
	if c.BlendBeta < 0 {
		errs = append(errs, ValidationError{Field: "blend_beta", Message: "must be >= 0"})
	}
	if c.BlendGamma < 0 {
		errs = append(errs, ValidationError{Field: "blend_gamma", Message: "must be >= 0"})
	}
	if c.ProfileBoost < 0 {
		errs = append(errs, ValidationError{Field: "profile_boost", Message: "must be >= 0"})
	}
	if c.ProfileWindowDays < 0 {
		errs = append(errs, ValidationError{Field: "profile_window_days", Message: "must be >= 0"})
	}
	if c.ProfileTopNTags < 0 {
		errs = append(errs, ValidationError{Field: "profile_top_n_tags", Message: "must be >= 0"})
	}
	if c.ProfileMinEventsForBoost < -1 {
		errs = append(errs, ValidationError{Field: "profile_min_events_for_boost", Message: "must be >= -1"})
	}
	if c.ProfileColdStartMultiplier < 0 || c.ProfileColdStartMultiplier > 1 {
		errs = append(errs, ValidationError{Field: "profile_cold_start_multiplier", Message: "must be between 0 and 1"})
	}
	if c.ProfileStarterBlendWeight < 0 || c.ProfileStarterBlendWeight > 1 {
		errs = append(errs, ValidationError{Field: "profile_starter_blend_weight", Message: "must be between 0 and 1"})
	}
	if c.MMRLambda < 0 || c.MMRLambda > 1 {
		errs = append(errs, ValidationError{Field: "mmr_lambda", Message: "must be between 0 and 1"})
	}
	if c.BrandCap < 0 {
		errs = append(errs, ValidationError{Field: "brand_cap", Message: "must be >= 0"})
	}
	if c.CategoryCap < 0 {
		errs = append(errs, ValidationError{Field: "category_cap", Message: "must be >= 0"})
	}
	if c.HalfLifeDays < 0 {
		errs = append(errs, ValidationError{Field: "half_life_days", Message: "must be >= 0"})
	}
	if c.CoVisWindowDays < 0 {
		errs = append(errs, ValidationError{Field: "co_vis_window_days", Message: "must be >= 0"})
	}
	if c.PurchasedWindowDays < 0 {
		errs = append(errs, ValidationError{Field: "purchased_window_days", Message: "must be >= 0"})
	}
	if c.PopularityFanout < 0 {
		errs = append(errs, ValidationError{Field: "popularity_fanout", Message: "must be >= 0"})
	}
	if c.MaxK < 0 {
		errs = append(errs, ValidationError{Field: "max_k", Message: "must be >= 0"})
	}
	if c.MaxFanout < 0 {
		errs = append(errs, ValidationError{Field: "max_fanout", Message: "must be >= 0"})
	}
	if c.MaxExcludeIDs < 0 {
		errs = append(errs, ValidationError{Field: "max_exclude_ids", Message: "must be >= 0"})
	}
	if c.MaxAnchorsInjected < 0 {
		errs = append(errs, ValidationError{Field: "max_anchors_injected", Message: "must be >= 0"})
	}
	if c.SessionLookbackEvents < 0 {
		errs = append(errs, ValidationError{Field: "session_lookback_events", Message: "must be >= 0"})
	}
	if c.SessionLookaheadMinutes < 0 {
		errs = append(errs, ValidationError{Field: "session_lookahead_minutes", Message: "must be >= 0"})
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Validate checks request invariants.
func (r Request) Validate() error {
	var errs ValidationErrors
	if r.OrgID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "org_id", Message: "must be set"})
	}
	if r.K < 0 {
		errs = append(errs, ValidationError{Field: "k", Message: "must be >= 0"})
	}
	if r.RecentEventCount < 0 {
		errs = append(errs, ValidationError{Field: "recent_event_count", Message: "must be >= 0"})
	}
	if r.StarterBlendWeight < 0 || r.StarterBlendWeight > 1 {
		errs = append(errs, ValidationError{Field: "starter_blend_weight", Message: "must be between 0 and 1"})
	}
	if r.Blend != nil {
		if r.Blend.Pop < 0 {
			errs = append(errs, ValidationError{Field: "blend.pop", Message: "must be >= 0"})
		}
		if r.Blend.Cooc < 0 {
			errs = append(errs, ValidationError{Field: "blend.cooc", Message: "must be >= 0"})
		}
		if r.Blend.Similarity < 0 {
			errs = append(errs, ValidationError{Field: "blend.similarity", Message: "must be >= 0"})
		}
	}
	if r.Constraints != nil {
		if r.Constraints.MinPrice != nil && r.Constraints.MaxPrice != nil {
			if *r.Constraints.MinPrice > *r.Constraints.MaxPrice {
				errs = append(errs, ValidationError{Field: "constraints.price", Message: "min_price must be <= max_price"})
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
