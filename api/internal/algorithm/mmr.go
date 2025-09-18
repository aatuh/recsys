package algorithm

import (
	"math"
	"strings"

	"recsys/internal/types"
)

// MMRReRankWithMetadata performs MMR re-ranking and returns per-item metadata.
func MMRReRankWithMetadata(
	candidates []types.ScoredItem,
	tags map[string]types.ItemTags,
	k int,
	lambda float64,
	brandCap, categoryCap int,
) ([]types.ScoredItem, map[string]MMRExplain, map[string]CapsExplain) {
	return mmrReRankInternal(candidates, tags, k, lambda, brandCap, categoryCap)
}

// MMRReRank performs MMR (Maximal Marginal Relevance) re-ranking on candidate
// items using tag overlap as a similarity proxy. It also enforces brand/category
// caps but discards metadata for compatibility.
func MMRReRank(
	candidates []types.ScoredItem,
	tags map[string]types.ItemTags,
	k int,
	lambda float64,
	brandCap, categoryCap int,
) []types.ScoredItem {
	items, _, _ := mmrReRankInternal(candidates, tags, k, lambda, brandCap, categoryCap)
	return items
}

func mmrReRankInternal(
	candidates []types.ScoredItem,
	tags map[string]types.ItemTags,
	k int,
	lambda float64,
	brandCap, categoryCap int,
) ([]types.ScoredItem, map[string]MMRExplain, map[string]CapsExplain) {
	if k <= 0 {
		k = 1
	}

	out := make([]types.ScoredItem, 0, minInt(k, len(candidates)))
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
	tagSets, brandMap, categoryMap := prepareTags(tags)

	// Track counts for caps.
	brandCount := make(map[string]int)
	categoryCount := make(map[string]int)
	selected := make(map[string]struct{})

	remainingCands := append([]types.ScoredItem(nil), candidates...)
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
				brandMap,
				categoryMap,
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
			if brand := brandMap[pick.ItemID]; brand != "" {
				usage.Applied = true
				count := brandCount[brand] + 1
				limit := brandCap
				usage.Count = &count
				usage.Limit = &limit
				usage.Value = brand
			}
			capsExplain.Brand = usage
		}
		if categoryCap > 0 {
			usage := &CapUsage{Applied: false}
			if category := categoryMap[pick.ItemID]; category != "" {
				usage.Applied = true
				count := categoryCount[category] + 1
				limit := categoryCap
				usage.Count = &count
				usage.Limit = &limit
				usage.Value = category
			}
			capsExplain.Category = usage
		}
		if capsExplain.Brand != nil || capsExplain.Category != nil {
			capsInfo[pick.ItemID] = capsExplain
		}

		// Update counts.
		if brand := brandMap[pick.ItemID]; brand != "" {
			brandCount[brand]++
		}
		if category := categoryMap[pick.ItemID]; category != "" {
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

// prepareTags extracts tag sets, brand, and category mappings from tags.
func prepareTags(tags map[string]types.ItemTags) (
	map[string]map[string]struct{},
	map[string]string,
	map[string]string,
) {
	tagSets := make(map[string]map[string]struct{})
	brandMap := make(map[string]string)
	categoryMap := make(map[string]string)

	for itemId, itemTags := range tags {
		if len(itemTags.Tags) == 0 {
			continue
		}

		tagSet := make(map[string]struct{})

		for _, tag := range itemTags.Tags {
			lowerTag := strings.ToLower(strings.TrimSpace(tag))

			switch {
			case strings.HasPrefix(lowerTag, "brand:"):
				brandMap[itemId] = strings.TrimSpace(lowerTag[len("brand:"):])
			case strings.HasPrefix(lowerTag, "category:"):
				categoryMap[itemId] = strings.TrimSpace(lowerTag[len("category:"):])
			case strings.HasPrefix(lowerTag, "cat:"):
				if _, ok := categoryMap[itemId]; !ok {
					categoryMap[itemId] = strings.TrimSpace(lowerTag[len("cat:"):])
				}
			default:
				if lowerTag != "" {
					tagSet[lowerTag] = struct{}{}
				}
			}
		}

		if len(tagSet) > 0 {
			tagSets[itemId] = tagSet
		}
	}

	return tagSets, brandMap, categoryMap
}

// canSelectWithCaps checks if an item can be selected given the current caps.
func canSelectWithCaps(
	itemId string,
	brandMap, categoryMap map[string]string,
	brandCount, categoryCount map[string]int,
	brandCap int,
	categoryCap int,
) bool {
	if brandCap > 0 {
		// Check if the brand count is greater than or equal to the brand cap.
		if brand := brandMap[itemId]; brand != "" &&
			brandCount[brand] >= brandCap {
			return false
		}
	}

	if categoryCap > 0 {
		// Check if the cat count is greater than or equal to the cat cap.
		if category := categoryMap[itemId]; category != "" &&
			categoryCount[category] >= categoryCap {
			return false
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
