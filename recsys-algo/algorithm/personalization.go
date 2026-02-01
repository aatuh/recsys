package algorithm

import (
	"context"

	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"
)

// applyPersonalizationBoost applies personalization boost based on user
// profile.
func (e *Engine) applyPersonalizationBoost(
	ctx context.Context, data *CandidateData, req Request,
) {
	if req.UserID == "" || e.config.ProfileBoost <= 0 {
		return
	}
	profileStore, ok := e.store.(recmodel.ProfileStore)
	if !ok {
		return
	}

	// Build the user tag profile. Profile is a map of tag:weight where weights
	// sum to 1.
	profile, err := profileStore.BuildUserTagProfile(
		ctx,
		req.OrgID,
		req.Namespace,
		req.UserID,
		e.config.ProfileWindowDays,
		maxInt(e.config.ProfileTopNTags, 1),
	)
	if err != nil {
		profile = nil
	} else {
		profile = normalizeTagWeights(profile)
	}
	if len(req.StarterProfile) > 0 {
		if len(profile) == 0 {
			profile = copyFloatMap(req.StarterProfile)
		} else if req.StarterBlendWeight > 0 {
			profile = blendTagProfiles(profile, req.StarterProfile, req.StarterBlendWeight)
		}
	}
	if len(profile) == 0 {
		return
	}

	anchorCount := effectiveAnchorCount(data.Anchors)
	eventCount := req.RecentEventCount
	if eventCount <= 0 {
		eventCount = anchorCount
	} else if anchorCount > 0 && eventCount > anchorCount {
		eventCount = anchorCount
	}
	minEvents := e.config.ProfileMinEventsForBoost
	if minEvents < 0 {
		minEvents = 0
	}
	coldScale := clampFloat(e.config.ProfileColdStartMultiplier, 0, 1)

	for i := range data.Candidates {
		itemId := data.Candidates[i].ItemID
		tags := data.Tags[itemId]

		overlap := 0.0
		for _, tag := range tags.Tags {
			if weight, ok := profile[tag]; ok {
				overlap += weight
			}
		}

		if overlap > 0 {
			multiplier := 1.0 + e.config.ProfileBoost*overlap
			attenuation := 1.0
			if minEvents > 0 {
				if eventCount == 0 || eventCount < minEvents {
					attenuation = coldScale
				}
			}
			if attenuation < 1.0 {
				multiplier = 1.0 + (multiplier-1.0)*attenuation
			}
			data.Candidates[i].Score *= multiplier
			data.Boosted[itemId] = true
			data.ProfileOverlap[itemId] = overlap
			data.ProfileMultiplier[itemId] = multiplier
		}
	}
}

func blendTagProfiles(primary, starter map[string]float64, starterWeight float64) map[string]float64 {
	weight := clampFloat(starterWeight, 0, 1)
	if weight == 0 {
		return copyFloatMap(primary)
	}
	if len(primary) == 0 {
		return copyFloatMap(starter)
	}

	blended := make(map[string]float64, len(primary)+len(starter))
	profileWeight := 1 - weight
	if profileWeight > 0 {
		for k, v := range primary {
			if v <= 0 {
				continue
			}
			blended[k] = v * profileWeight
		}
	}
	if weight > 0 {
		for k, v := range starter {
			if v <= 0 {
				continue
			}
			blended[k] += v * weight
		}
	}
	normalizeFloatMap(blended)
	return blended
}

func normalizeTagWeights(profile map[string]float64) map[string]float64 {
	if len(profile) == 0 {
		return nil
	}
	normalized := make(map[string]float64, len(profile))
	for tag, weight := range profile {
		if weight <= 0 {
			continue
		}
		key := recmodel.NormalizeTag(tag)
		if key == "" {
			continue
		}
		normalized[key] += weight
	}
	normalizeFloatMap(normalized)
	return normalized
}

func normalizeFloatMap(values map[string]float64) {
	if len(values) == 0 {
		return
	}
	sum := 0.0
	for _, v := range values {
		if v > 0 {
			sum += v
		}
	}
	if sum <= 0 {
		for k := range values {
			delete(values, k)
		}
		return
	}
	for k, v := range values {
		if v <= 0 {
			delete(values, k)
			continue
		}
		values[k] = v / sum
	}
}

func effectiveAnchorCount(anchors []string) int {
	if len(anchors) == 0 {
		return 0
	}
	if len(anchors) == 1 && anchors[0] == "(no_recent_activity)" {
		return 0
	}
	return len(anchors)
}
