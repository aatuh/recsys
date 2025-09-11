package algorithm

import (
	"math"
	"strings"

	"recsys/internal/types"
)

// MMRReRank performs MMR (Maximal Marginal Relevance) re-ranking on candidate items
// using tag overlap as a similarity proxy. It also enforces brand/category caps.
//
// Parameters:
//   - candidates: Items to re-rank
//   - meta: Item metadata for diversity calculation
//   - k: Number of items to select
//   - lambda: MMR balance (0=diversity, 1=relevance)
//   - brandCap: Max items per brand (0=disabled)
//   - categoryCap: Max items per category (0=disabled)
//
// Returns: Re-ranked items up to k items
func MMRReRank(
	candidates []types.ScoredItem,
	meta map[string]types.ItemMeta,
	k int,
	lambda float64,
	brandCap, categoryCap int,
) []types.ScoredItem {
	if k <= 0 {
		k = 1
	}
	
	out := make([]types.ScoredItem, 0, minInt(k, len(candidates)))
	if len(candidates) == 0 {
		return out
	}

	// Precompute normalized scores
	maxScore := 0.0
	for _, candidate := range candidates {
		if candidate.Score > maxScore {
			maxScore = candidate.Score
		}
	}
	
	normScore := func(s float64) float64 {
		if maxScore <= 0 {
			return 0
		}
		return s / maxScore
	}

	// Prepare metadata for efficient lookup
	tagSets, brandMap, categoryMap := prepareMetadata(meta)

	// Track counts for caps
	brandCount := make(map[string]int)
	categoryCount := make(map[string]int)
	selected := make(map[string]struct{})

	remaining := append([]types.ScoredItem(nil), candidates...)

	// Greedy MMR with deterministic tie-break by initial order
	for len(out) < k && len(remaining) > 0 {
		bestIdx := -1
		bestMMR := math.Inf(-1)

		for i, candidate := range remaining {
			id := candidate.ItemID

			// Enforce caps
			if !canSelectWithCaps(id, brandMap, categoryMap, brandCount, categoryCount, brandCap, categoryCap) {
				continue
			}

			// Calculate diversity term: max similarity to already selected
			maxSim := 0.0
			if len(selected) > 0 {
				for selectedID := range selected {
					sim := jaccard(tagSets[id], tagSets[selectedID])
					if sim > maxSim {
						maxSim = sim
					}
				}
			}

			// MMR score: lambda * relevance - (1-lambda) * diversity
			score := lambda*normScore(candidate.Score) - (1.0-lambda)*maxSim
			if score > bestMMR {
				bestMMR = score
				bestIdx = i
			}
		}

		// If caps block all remaining items, stop selection early
		if bestIdx == -1 {
			break
		}

		pick := remaining[bestIdx]
		out = append(out, pick)
		selected[pick.ItemID] = struct{}{}
		
		// Update counts
		if brand := brandMap[pick.ItemID]; brand != "" {
			brandCount[brand]++
		}
		if category := categoryMap[pick.ItemID]; category != "" {
			categoryCount[category]++
		}

		// Remove selected item from remaining
		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}

	return out
}

// prepareMetadata extracts tag sets, brand, and category mappings from metadata
func prepareMetadata(meta map[string]types.ItemMeta) (
	tagSets map[string]map[string]struct{},
	brandMap map[string]string,
	categoryMap map[string]string,
) {
	tagSets = make(map[string]map[string]struct{})
	brandMap = make(map[string]string)
	categoryMap = make(map[string]string)

	for id, itemMeta := range meta {
		if len(itemMeta.Tags) == 0 {
			continue
		}

		tagSet := make(map[string]struct{})
		
		for _, tag := range itemMeta.Tags {
			lowerTag := strings.ToLower(strings.TrimSpace(tag))
			
			switch {
			case strings.HasPrefix(lowerTag, "brand:"):
				brandMap[id] = strings.TrimSpace(lowerTag[len("brand:"):])
			case strings.HasPrefix(lowerTag, "category:"):
				categoryMap[id] = strings.TrimSpace(lowerTag[len("category:"):])
			case strings.HasPrefix(lowerTag, "cat:"):
				if _, ok := categoryMap[id]; !ok {
					categoryMap[id] = strings.TrimSpace(lowerTag[len("cat:"):])
				}
			default:
				if lowerTag != "" {
					tagSet[lowerTag] = struct{}{}
				}
			}
		}
		
		if len(tagSet) > 0 {
			tagSets[id] = tagSet
		}
	}

	return tagSets, brandMap, categoryMap
}

// canSelectWithCaps checks if an item can be selected given the current caps
func canSelectWithCaps(
	id string,
	brandMap, categoryMap map[string]string,
	brandCount, categoryCount map[string]int,
	brandCap, categoryCap int,
) bool {
	if brandCap > 0 {
		if brand := brandMap[id]; brand != "" && brandCount[brand] >= brandCap {
			return false
		}
	}
	
	if categoryCap > 0 {
		if category := categoryMap[id]; category != "" && categoryCount[category] >= categoryCap {
			return false
		}
	}
	
	return true
}

// jaccard computes Jaccard similarity of two string sets
func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	
	intersection := 0
	for k := range a {
		if _, ok := b[k]; ok {
			intersection++
		}
	}
	
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	
	return float64(intersection) / float64(union)
}
