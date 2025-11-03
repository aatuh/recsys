package recommendation

import (
	"context"
	"sort"
	"strings"
	"time"

	"recsys/internal/algorithm"
)

var starterCategoryPresets = map[string]map[string]float64{
	"new_users": {
		"electronics": 0.25,
		"books":       0.2,
		"home":        0.2,
		"fashion":     0.2,
		"beauty":      0.15,
	},
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
) map[string]float64 {
	if req.UserID == "" {
		return nil
	}

	if s.userHasHistory(ctx, cfg, req) {
		return nil
	}

	segmentID := strings.ToLower(strings.TrimSpace(selection.SegmentID))
	if segmentID == "" && selection.UserTraits != nil {
		if seg, ok := selection.UserTraits["segment"].(string); ok {
			segmentID = strings.ToLower(strings.TrimSpace(seg))
		}
	}
	if segmentID == "" {
		return nil
	}

	tagProfile := starterTagProfileForSegment(segmentID)
	if len(tagProfile) == 0 {
		return nil
	}

	topN := cfg.ProfileTopNTags
	if topN <= 0 {
		topN = len(tagProfile)
	}
	tagProfile = trimTagProfile(tagProfile, topN)
	return tagProfile
}

func (s *Service) userHasHistory(ctx context.Context, cfg algorithm.Config, req algorithm.Request) bool {
	if s.store == nil || req.UserID == "" {
		return false
	}
	days := cfg.ProfileWindowDays
	if days <= 0 {
		days = 30
	}
	since := time.Now().UTC().Add(-time.Duration(days*24.0) * time.Hour)
	items, err := s.store.ListUserRecentItemIDs(ctx, req.OrgID, req.Namespace, req.UserID, since, 1)
	if err != nil {
		return true
	}
	return len(items) > 0
}

func starterTagProfileForSegment(segment string) map[string]float64 {
	preset, ok := starterCategoryPresets[segment]
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
