package recommendation

import (
	"context"
	"sort"
	"strings"
	"time"

	"recsys/internal/algorithm"
)

func defaultStarterPresets() map[string]map[string]float64 {
	return map[string]map[string]float64{
		"new_users": {
			"electronics": 0.25,
			"books":       0.2,
			"home":        0.2,
			"fashion":     0.2,
			"beauty":      0.15,
		},
	}
}

var categoryTagMapping = map[string][]string{
	"electronics": {"electronics", "gadgets", "smart"},
	"books":       {"books", "reading", "literature"},
	"home":        {"home", "decor", "living"},
	"fashion":     {"fashion", "style", "apparel"},
	"beauty":      {"beauty", "skincare", "wellness"},
	"gourmet":     {"food", "gourmet", "kitchen"},
	"fitness":     {"fitness", "health", "active"},
	"outdoors":    {"outdoors", "adventure", "travel"},
}

func (s *Service) buildStarterProfile(
	ctx context.Context,
	cfg algorithm.Config,
	req algorithm.Request,
	selection SegmentSelection,
	recentEventCount int,
	recentCountKnown bool,
	recentItemIDs []string,
) (map[string]float64, float64) {
	if req.UserID == "" {
		return nil, 0
	}

	segmentID := strings.ToLower(strings.TrimSpace(selection.SegmentID))
	if segmentID == "" && selection.UserTraits != nil {
		if seg, ok := selection.UserTraits["segment"].(string); ok {
			segmentID = strings.ToLower(strings.TrimSpace(seg))
		}
	}
	if segmentID == "" {
		segmentID = "new_users"
	}

	isNewSegment := segmentID == "new_users"

	minEvents := cfg.ProfileMinEventsForBoost
	if minEvents < 0 {
		minEvents = 0
	}
	isSparseHistory := false
	if !recentCountKnown {
		isSparseHistory = true
	} else if minEvents == 0 {
		isSparseHistory = recentEventCount == 0
	} else {
		isSparseHistory = recentEventCount < minEvents
	}

	if !isNewSegment && !isSparseHistory {
		return nil, 0
	}

	tagProfile := starterTagProfileForSegment(segmentID, s.starterPresets)
	if len(tagProfile) == 0 && segmentID != "new_users" {
		tagProfile = starterTagProfileForSegment("new_users", s.starterPresets)
	}
	if len(tagProfile) == 0 {
		return nil, 0
	}

	blendWeight := cfg.ProfileStarterBlendWeight
	if blendWeight < 0 {
		blendWeight = 0
	} else if blendWeight > 1 {
		blendWeight = 1
	}

	decayEvents := s.starterDecayEvents
	if decayEvents <= 0 {
		decayEvents = minEvents
	}
	if decayEvents <= 0 {
		decayEvents = 3
	}
	starterWeight := blendWeight
	missing := decayEvents - recentEventCount
	if !recentCountKnown {
		missing = decayEvents
	}
	if missing < 0 {
		missing = 0
	}
	factor := float64(missing) / float64(decayEvents)
	if factor > 1 {
		factor = 1
	}
	starterWeight = starterWeight + (1-starterWeight)*factor
	if starterWeight < 0 {
		starterWeight = 0
	} else if starterWeight > 1 {
		starterWeight = 1
	}

	if len(recentItemIDs) > 0 {
		tagsByItem, err := s.store.ListItemsTags(ctx, req.OrgID, req.Namespace, recentItemIDs)
		if err == nil && len(tagsByItem) > 0 {
			recentProfile := make(map[string]float64)
			for _, tags := range tagsByItem {
				if len(tags.Tags) == 0 {
					continue
				}
				weight := 1.0 / float64(len(tags.Tags))
				for _, tag := range tags.Tags {
					key := strings.ToLower(strings.TrimSpace(tag))
					if key == "" {
						continue
					}
					recentProfile[key] += weight
				}
			}
			normalizeWeights(recentProfile)
			if len(recentProfile) > 0 {
				tagProfile = blendProfiles(tagProfile, recentProfile, 1-starterWeight)
			}
		}
	}

	topN := cfg.ProfileTopNTags
	if topN <= 0 {
		topN = len(tagProfile)
	}
	tagProfile = trimTagProfile(tagProfile, topN)
	return tagProfile, starterWeight
}

func (s *Service) recentInteractionItems(ctx context.Context, cfg algorithm.Config, req algorithm.Request) ([]string, error) {
	if s.store == nil || req.UserID == "" {
		return nil, nil
	}
	days := cfg.ProfileWindowDays
	if days <= 0 {
		days = 30
	}
	since := time.Now().UTC().Add(-time.Duration(days*24.0) * time.Hour)
	limit := cfg.ProfileMinEventsForBoost + 1
	if limit <= 0 {
		limit = 1
	}
	if limit < 10 {
		limit = 10
	}
	items, err := s.store.ListUserRecentItemIDs(ctx, req.OrgID, req.Namespace, req.UserID, since, limit)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func starterTagProfileForSegment(segment string, presets map[string]map[string]float64) map[string]float64 {
	if len(presets) == 0 {
		presets = defaultStarterPresets()
	}
	preset, ok := presets[segment]
	if !ok {
		return nil
	}
	tagWeights := make(map[string]float64)
	for category, weight := range preset {
		tags := categoryTagMapping[strings.ToLower(strings.TrimSpace(category))]
		if len(tags) == 0 || weight <= 0 {
			continue
		}
		perTag := weight / float64(len(tags))
		for _, tag := range tags {
			key := strings.ToLower(strings.TrimSpace(tag))
			if key == "" {
				continue
			}
			tagWeights[key] += perTag
		}
	}
	total := 0.0
	for _, v := range tagWeights {
		total += v
	}
	if total == 0 {
		return nil
	}
	normalized := make(map[string]float64, len(tagWeights))
	for k, v := range tagWeights {
		normalized[k] = v / total
	}
	return normalized
}

func trimTagProfile(profile map[string]float64, topN int) map[string]float64 {
	if topN <= 0 || len(profile) <= topN {
		return profile
	}
	type kv struct {
		Key   string
		Value float64
	}
	pairs := make([]kv, 0, len(profile))
	for k, v := range profile {
		pairs = append(pairs, kv{Key: k, Value: v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[j].Value
	})
	if topN > len(pairs) {
		topN = len(pairs)
	}
	selected := make(map[string]float64, topN)
	total := 0.0
	for i := 0; i < topN; i++ {
		selected[pairs[i].Key] = pairs[i].Value
		total += pairs[i].Value
	}
	if total == 0 {
		return selected
	}
	for k, v := range selected {
		selected[k] = v / total
	}
	return selected
}

func normalizeWeights(values map[string]float64) {
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

func blendProfiles(primary, secondary map[string]float64, weight float64) map[string]float64 {
	w := weight
	if w < 0 {
		w = 0
	}
	if w > 1 {
		w = 1
	}
	if w == 0 {
		return copyWeights(primary)
	}
	if len(primary) == 0 {
		return copyWeights(secondary)
	}
	blended := make(map[string]float64, len(primary)+len(secondary))
	primaryWeight := 1 - w
	if primaryWeight > 0 {
		for k, v := range primary {
			if v <= 0 {
				continue
			}
			blended[k] = v * primaryWeight
		}
	}
	if w > 0 {
		for k, v := range secondary {
			if v <= 0 {
				continue
			}
			blended[k] += v * w
		}
	}
	normalizeWeights(blended)
	return blended
}

func copyWeights(src map[string]float64) map[string]float64 {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]float64, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}
