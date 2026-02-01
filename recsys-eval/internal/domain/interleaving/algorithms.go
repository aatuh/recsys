package interleaving

import (
	"math/rand"
)

// Result is an interleaved list with attribution per position.
type Result struct {
	Items       []string
	Attribution []string // "A" or "B"
}

// TeamDraft interleaves by alternating picks between rankers.
func TeamDraft(listA, listB []string, maxResults int, rng *rand.Rand) Result {
	maxResults = normalizeMax(maxResults, listA, listB)
	items := make([]string, 0, maxResults)
	attr := make([]string, 0, maxResults)
	seen := map[string]struct{}{}

	turnA := rng == nil || rng.Intn(2) == 0

	idxA := 0
	idxB := 0
	for len(items) < maxResults && (idxA < len(listA) || idxB < len(listB)) {
		if turnA {
			pick(&items, &attr, &seen, "A", listA, &idxA, maxResults)
		} else {
			pick(&items, &attr, &seen, "B", listB, &idxB, maxResults)
		}
		turnA = !turnA
	}

	return Result{Items: items, Attribution: attr}
}

// BalancedInterleaving ensures equal contribution from each ranker.
func BalancedInterleaving(listA, listB []string, maxResults int, rng *rand.Rand) Result {
	maxResults = normalizeMax(maxResults, listA, listB)
	items := make([]string, 0, maxResults)
	attr := make([]string, 0, maxResults)
	seen := map[string]struct{}{}
	idxA := 0
	idxB := 0
	countA := 0
	countB := 0

	for len(items) < maxResults && (idxA < len(listA) || idxB < len(listB)) {
		chooseA := countA <= countB
		if countA == countB && rng != nil && rng.Intn(2) == 1 {
			chooseA = false
		}
		if chooseA {
			picked := pick(&items, &attr, &seen, "A", listA, &idxA, maxResults)
			if picked {
				countA++
			}
		} else {
			picked := pick(&items, &attr, &seen, "B", listB, &idxB, maxResults)
			if picked {
				countB++
			}
		}
	}

	return Result{Items: items, Attribution: attr}
}

// OptimizedInterleaving favors unique items to increase sensitivity.
func OptimizedInterleaving(listA, listB []string, maxResults int, rng *rand.Rand) Result {
	maxResults = normalizeMax(maxResults, listA, listB)
	items := make([]string, 0, maxResults)
	attr := make([]string, 0, maxResults)
	seen := map[string]struct{}{}
	idxA := 0
	idxB := 0
	countA := 0
	countB := 0

	for len(items) < maxResults && (idxA < len(listA) || idxB < len(listB)) {
		okA := hasUnique(listA, idxA, seen)
		okB := hasUnique(listB, idxB, seen)

		var chooseA bool
		switch {
		case okA && !okB:
			chooseA = true
		case !okA && okB:
			chooseA = false
		case okA && okB:
			// Prefer the list with fewer contributions to keep balance.
			if countA == countB {
				chooseA = true
				if rng != nil && rng.Intn(2) == 1 {
					chooseA = false
				}
			} else {
				chooseA = countA < countB
			}
		default:
			// Fall back to balanced when only overlaps remain.
			chooseA = countA <= countB
		}

		if chooseA {
			picked := pick(&items, &attr, &seen, "A", listA, &idxA, maxResults)
			if picked {
				countA++
			}
		} else {
			picked := pick(&items, &attr, &seen, "B", listB, &idxB, maxResults)
			if picked {
				countB++
			}
		}
	}

	return Result{Items: items, Attribution: attr}
}

func normalizeMax(maxResults int, listA, listB []string) int {
	if maxResults <= 0 {
		if len(listA) > len(listB) {
			return len(listA)
		}
		return len(listB)
	}
	return maxResults
}

func pick(items *[]string, attr *[]string, seen *map[string]struct{}, team string, list []string, idx *int, maxResults int) bool {
	for *idx < len(list) && len(*items) < maxResults {
		item := list[*idx]
		*idx = *idx + 1
		if _, ok := (*seen)[item]; ok {
			continue
		}
		(*seen)[item] = struct{}{}
		*items = append(*items, item)
		*attr = append(*attr, team)
		return true
	}
	return false
}

func hasUnique(list []string, idx int, seen map[string]struct{}) bool {
	for i := idx; i < len(list); i++ {
		item := list[i]
		if _, ok := seen[item]; !ok {
			return true
		}
	}
	return false
}
