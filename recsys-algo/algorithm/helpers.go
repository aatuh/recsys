package algorithm

import (
	"strings"

	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"
)

func (e *Engine) sanitizeRequest(req Request) Request {
	k := req.K
	if k <= 0 {
		k = 20
	}
	if e.config.MaxK > 0 && k > e.config.MaxK {
		k = e.config.MaxK
	}
	req.K = k

	if req.Constraints != nil {
		clone := *req.Constraints
		if len(clone.IncludeTagsAny) > 0 {
			clone.IncludeTagsAny = append([]string(nil), clone.IncludeTagsAny...)
		}
		if len(clone.ExcludeItemIDs) > 0 {
			exclude := clone.ExcludeItemIDs
			if e.config.MaxExcludeIDs > 0 && len(exclude) > e.config.MaxExcludeIDs {
				exclude = exclude[:e.config.MaxExcludeIDs]
			}
			clone.ExcludeItemIDs = append([]string(nil), exclude...)
		}
		req.Constraints = &clone
	}

	if req.InjectAnchors && len(req.AnchorItemIDs) > 0 {
		req.AnchorItemIDs = clampAnchorIDs(req.AnchorItemIDs, e.config.MaxAnchorsInjected)
	}

	if len(req.StarterProfile) > 0 {
		req.StarterProfile = normalizeTagWeights(req.StarterProfile)
	}

	return req
}

func resolveAlgorithm(request AlgorithmKind, fallback AlgorithmKind) AlgorithmKind {
	if request != "" {
		return NormalizeAlgorithm(request)
	}
	if fallback != "" {
		return NormalizeAlgorithm(fallback)
	}
	return AlgorithmBlend
}

func (e *Engine) fanoutFor(k int) int {
	fanout := e.config.PopularityFanout
	if fanout <= 0 || fanout < k {
		fanout = k
	}
	if e.config.MaxFanout > 0 && fanout > e.config.MaxFanout {
		fanout = e.config.MaxFanout
	}
	if fanout < 1 {
		fanout = k
	}
	return fanout
}

func clampAnchorIDs(anchorIDs []string, limit int) []string {
	if len(anchorIDs) == 0 {
		return nil
	}
	out := make([]string, 0, len(anchorIDs))
	seen := make(map[string]struct{}, len(anchorIDs))
	for _, anchor := range anchorIDs {
		if limit > 0 && len(out) >= limit {
			break
		}
		trimmed := strings.TrimSpace(anchor)
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

func normalizeItemTags(tags map[string]recmodel.ItemTags) map[string]recmodel.ItemTags {
	if len(tags) == 0 {
		return tags
	}
	normalized := make(map[string]recmodel.ItemTags, len(tags))
	for id, info := range tags {
		info.Tags = recmodel.NormalizeTags(info.Tags)
		normalized[id] = info
	}
	return normalized
}

func copyScoredItems(items []recmodel.ScoredItem) []recmodel.ScoredItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]recmodel.ScoredItem, len(items))
	copy(out, items)
	return out
}

func copyMMRInfo(src map[string]MMRExplain) map[string]MMRExplain {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]MMRExplain, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func copyCapsInfo(src map[string]CapsExplain) map[string]CapsExplain {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]CapsExplain, len(src))
	for k, v := range src {
		out[k] = CapsExplain{
			Brand:    copyCapUsage(v.Brand),
			Category: copyCapUsage(v.Category),
		}
	}
	return out
}

func copyCapUsage(src *CapUsage) *CapUsage {
	if src == nil {
		return nil
	}
	usage := &CapUsage{Applied: src.Applied, Value: src.Value}
	if src.Limit != nil {
		limit := *src.Limit
		usage.Limit = &limit
	}
	if src.Count != nil {
		count := *src.Count
		usage.Count = &count
	}
	return usage
}

func copyBoolMap(src map[string]bool) map[string]bool {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]bool, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func copySignalStatus(src map[Signal]SignalStatus) map[Signal]SignalStatus {
	if len(src) == 0 {
		return nil
	}
	out := make(map[Signal]SignalStatus, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func copyFloatMap(src map[string]float64) map[string]float64 {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]float64, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func clampFloat(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// maxInt returns the maximum of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
