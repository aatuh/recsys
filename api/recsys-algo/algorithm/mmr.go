package algorithm

import (
	"math"
	"strings"

	recmodel "github.com/aatuh/recsys-algo/model"
)

var (
	defaultBrandTagPrefixes    = []string{"brand"}
	defaultCategoryTagPrefixes = []string{"category", "cat"}
)

// MMRReRankWithMetadata performs MMR re-ranking and returns per-item metadata.
func MMRReRankWithMetadata(
	candidates []recmodel.ScoredItem,
	tags map[string]recmodel.ItemTags,
	k int,
	lambda float64,
	brandCap, categoryCap int,
) ([]recmodel.ScoredItem, map[string]MMRExplain, map[string]CapsExplain) {
	return mmrReRankInternal(
		candidates,
		tags,
		k,
		lambda,
		brandCap,
		categoryCap,
		defaultBrandTagPrefixes,
		defaultCategoryTagPrefixes,
	)
}

// MMRReRank performs MMR (Maximal Marginal Relevance) re-ranking on candidate
// items using tag overlap as a similarity proxy. It also enforces brand/category
// caps but discards metadata for compatibility.
func MMRReRank(
	candidates []recmodel.ScoredItem,
	tags map[string]recmodel.ItemTags,
	k int,
	lambda float64,
	brandCap, categoryCap int,
) []recmodel.ScoredItem {
	items, _, _ := mmrReRankInternal(
		candidates,
		tags,
		k,
		lambda,
		brandCap,
		categoryCap,
		defaultBrandTagPrefixes,
		defaultCategoryTagPrefixes,
	)
	return items
}

func mmrReRankInternal(
	candidates []recmodel.ScoredItem,
	tags map[string]recmodel.ItemTags,
	k int,
	lambda float64,
	brandCap, categoryCap int,
	brandTagPrefixes, categoryTagPrefixes []string,
) ([]recmodel.ScoredItem, map[string]MMRExplain, map[string]CapsExplain) {
	if k <= 0 {
		k = 1
	}

	out := make([]recmodel.ScoredItem, 0, minInt(k, len(candidates)))
	if len(candidates) == 0 {
		return out, map[string]MMRExplain{}, map[string]CapsExplain{}
	}

	mmrInfo := make(map[string]MMRExplain)
	capsInfo := make(map[string]CapsExplain)

	// Precompute normalized scores.
	maxScore := 0.0
	for _, candidate := range candidates {
		if candidate.Score > maxScore {
			maxScore = candidate.Score
		}
	}
	// Prepare metadata for efficient lookup.
	tagSets, brandValues, categoryValues := prepareTags(tags, brandTagPrefixes, categoryTagPrefixes)

	// Track counts for caps.
	brandCount := make(map[string]int)
	categoryCount := make(map[string]int)
	selected := make(map[string]struct{})

	remainingCands := append([]recmodel.ScoredItem(nil), candidates...)
	rank := 0

	// Greedy MMR with deterministic tie-break by initial order.
	for len(out) < k && len(remainingCands) > 0 {
		bestMMR := math.Inf(-1)
		bestIdx := -1
		bestMaxSim := 0.0
		bestNorm := 0.0

		// Iterate over the remaining candidates and find the best MMR score.
		for i, candidate := range remainingCands {
			itemId := candidate.ItemID

			// Enforce caps.
			if !canSelectWithCaps(
				itemId,
				brandValues,
				categoryValues,
				brandCount,
				categoryCount,
				brandCap,
				categoryCap,
			) {
				continue
			}

			// Calculate diversity term: max similarity to already selected.
			maxSim := 0.0
			if len(selected) > 0 {
				for selectedID := range selected {
					sim := jaccard(tagSets[itemId], tagSets[selectedID])
					if sim > maxSim {
						maxSim = sim
					}
				}
			}

			norm := normScore(candidate.Score, maxScore)
			score := lambda*norm - (1.0-lambda)*maxSim

			if score > bestMMR {
				bestMMR = score
				bestIdx = i
				bestMaxSim = maxSim
				bestNorm = norm
			}
		}

		// If caps block all remaining items, stop selection early.
		if bestIdx == -1 {
			break
		}

		pick := remainingCands[bestIdx]
		out = append(out, pick)
		selected[pick.ItemID] = struct{}{}
		rank++

		mmrInfo[pick.ItemID] = MMRExplain{
			Lambda:        lambda,
			MaxSimilarity: bestMaxSim,
			Penalty:       (1.0 - lambda) * bestMaxSim,
			Relevance:     lambda * bestNorm,
			Rank:          rank,
		}

		capsExplain := CapsExplain{}
		if brandCap > 0 {
			usage := &CapUsage{Applied: false}
			if brands := brandValues[pick.ItemID]; len(brands) > 0 {
				usage.Applied = true
				limit := brandCap
				usage.Limit = &limit
				usage.Value = strings.Join(brands, ",")
				if len(brands) == 1 {
					count := brandCount[brands[0]] + 1
					usage.Count = &count
				}
			}
			capsExplain.Brand = usage
		}
		if categoryCap > 0 {
			usage := &CapUsage{Applied: false}
			if categories := categoryValues[pick.ItemID]; len(categories) > 0 {
				usage.Applied = true
				limit := categoryCap
				usage.Limit = &limit
				usage.Value = strings.Join(categories, ",")
				if len(categories) == 1 {
					count := categoryCount[categories[0]] + 1
					usage.Count = &count
				}
			}
			capsExplain.Category = usage
		}
		if capsExplain.Brand != nil || capsExplain.Category != nil {
			capsInfo[pick.ItemID] = capsExplain
		}

		// Update counts.
		for _, brand := range brandValues[pick.ItemID] {
			brandCount[brand]++
		}
		for _, category := range categoryValues[pick.ItemID] {
			categoryCount[category]++
		}

		// Remove selected item from remaining.
		remainingCands = append(
			remainingCands[:bestIdx], remainingCands[bestIdx+1:]...,
		)
	}

	return out, mmrInfo, capsInfo
}

// normScore normalizes a score to [0, 1].
func normScore(score float64, maxScore float64) float64 {
	if maxScore <= 0 {
		return 0
	}
	return score / maxScore
}

// prepareTags extracts tag sets and structured tag values (brand/category) using
// configurable prefixes. Returned structured maps contain normalized values per
// item and may hold multiple entries when the data includes duplicates.
func prepareTags(
	tags map[string]recmodel.ItemTags,
	brandPrefixes, categoryPrefixes []string,
) (
	map[string]map[string]struct{},
	map[string][]string,
	map[string][]string,
) {
	toMatchers := func(prefixes []string) []string {
		if len(prefixes) == 0 {
			return nil
		}
		seen := make(map[string]struct{}, len(prefixes))
		out := make([]string, 0, len(prefixes))
		for _, raw := range prefixes {
			trimmed := recmodel.NormalizeTag(raw)
			trimmed = strings.TrimSuffix(trimmed, ":")
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			out = append(out, trimmed+":")
		}
		return out
	}

	matchStructured := func(tag string, matchers []string) (string, bool) {
		for _, prefix := range matchers {
			if strings.HasPrefix(tag, prefix) {
				val := strings.TrimSpace(tag[len(prefix):])
				if val != "" {
					return val, true
				}
			}
		}
		return "", false
	}

	appendUnique := func(dst map[string][]string, itemID, value string) {
		slice := dst[itemID]
		for _, existing := range slice {
			if existing == value {
				return
			}
		}
		dst[itemID] = append(slice, value)
	}

	brandMatchers := toMatchers(brandPrefixes)
	categoryMatchers := toMatchers(categoryPrefixes)
	tagSets := make(map[string]map[string]struct{})
	brandValues := make(map[string][]string)
	categoryValues := make(map[string][]string)

	for itemID, itemTags := range tags {
		if len(itemTags.Tags) == 0 {
			continue
		}

		tagSet := make(map[string]struct{})
		for _, tag := range itemTags.Tags {
			lowerTag := recmodel.NormalizeTag(tag)
			if lowerTag == "" {
				continue
			}

			if val, ok := matchStructured(lowerTag, brandMatchers); ok {
				appendUnique(brandValues, itemID, val)
				continue
			}
			if val, ok := matchStructured(lowerTag, categoryMatchers); ok {
				appendUnique(categoryValues, itemID, val)
				continue
			}

			tagSet[lowerTag] = struct{}{}
		}

		if len(tagSet) > 0 {
			tagSets[itemID] = tagSet
		}
	}

	return tagSets, brandValues, categoryValues
}

// canSelectWithCaps checks if an item can be selected given the current caps.
func canSelectWithCaps(
	itemId string,
	brandValues, categoryValues map[string][]string,
	brandCount, categoryCount map[string]int,
	brandCap int,
	categoryCap int,
) bool {
	if brandCap > 0 {
		for _, brand := range brandValues[itemId] {
			if brandCount[brand] >= brandCap {
				return false
			}
		}
	}

	if categoryCap > 0 {
		for _, category := range categoryValues[itemId] {
			if categoryCount[category] >= categoryCap {
				return false
			}
		}
	}

	return true
}

// jaccard computes Jaccard similarity of two string sets. It returns a value
// between 0 and 1.
func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	// Calculate the intersection of the two sets.
	intersection := 0
	for k := range a {
		if _, ok := b[k]; ok {
			intersection++
		}
	}

	// Calculate the union of the two sets.
	// Union is total unique elements: |A| + |B| - |A ∩ B|
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}
