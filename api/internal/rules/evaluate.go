package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"recsys/internal/types"
)

type evaluator struct {
	maxPinSlots         int
	brandTagPrefixes    []string
	categoryTagPrefixes []string
}

type itemState struct {
	blocked      bool
	blockRules   []uuid.UUID
	pinned       bool
	pinRules     []uuid.UUID
	boostDelta   float64
	boostRules   []BoostDetail
	pinPriority  int
	highestBoost BoostDetail
}

func (e *evaluator) apply(rules []types.Rule, req EvaluateRequest) (*EvaluateResult, error) {
	result := &EvaluateResult{
		Candidates:       nil,
		Pinned:           nil,
		Matches:          nil,
		EvaluatedRuleIDs: nil,
		ItemEffects:      make(map[string]ItemEffect),
		ReasonTags:       make(map[string][]string),
	}

	if len(req.Candidates) == 0 {
		result.Candidates = []types.ScoredItem{}
	}

	candidateMap := make(map[string]types.ScoredItem, len(req.Candidates))
	order := make([]string, 0, len(req.Candidates))
	for _, cand := range req.Candidates {
		candidateMap[cand.ItemID] = cand
		order = append(order, cand.ItemID)
	}

	tagSets, brandValues, categoryValues := e.prepareIndexes(req.ItemTags)

	states := make(map[string]*itemState)
	pinnedOrder := make([]string, 0)
	pinnedSeen := make(map[string]struct{})
	remainingPins := e.maxPinSlots

	matches := make([]Match, 0, len(rules))
	evaluated := make([]uuid.UUID, 0, len(rules))

	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].Priority == rules[j].Priority {
			return i < j
		}
		return rules[i].Priority > rules[j].Priority
	})

	for _, rule := range rules {
		evaluated = append(evaluated, rule.RuleID)

		matched := e.matchRule(rule, candidateMap, tagSets, brandValues, categoryValues)
		if len(matched) == 0 {
			continue
		}

		// Deduplicate item IDs for the match payload while preserving order.
		matched = deduplicatePreserveOrder(matched)
		matches = append(matches, Match{
			RuleID:  rule.RuleID,
			Action:  rule.Action,
			Target:  rule.TargetType,
			ItemIDs: append([]string(nil), matched...),
		})

		switch rule.Action {
		case types.RuleActionBlock:
			for _, id := range matched {
				st := ensureState(states, id)
				// Release pin slot if this item was pinned earlier.
				if st.pinned {
					st.pinned = false
					if remainingPins < e.maxPinSlots {
						remainingPins++
					}
				}
				st.blocked = true
				if !containsUUID(st.blockRules, rule.RuleID) {
					st.blockRules = append(st.blockRules, rule.RuleID)
				}
			}

		case types.RuleActionPin:
			if remainingPins == 0 {
				continue
			}
			perRuleLimit := e.maxPinSlots
			if rule.MaxPins != nil && *rule.MaxPins >= 0 && *rule.MaxPins < perRuleLimit {
				perRuleLimit = *rule.MaxPins
			}
			if perRuleLimit <= 0 {
				continue
			}
			rulePinned := 0
			for _, id := range matched {
				if remainingPins == 0 {
					break
				}
				if rule.MaxPins != nil && rulePinned >= perRuleLimit {
					break
				}
				st := ensureState(states, id)
				if st.blocked || st.pinned {
					continue
				}
				st.pinned = true
				st.pinPriority = rule.Priority
				if !containsUUID(st.pinRules, rule.RuleID) {
					st.pinRules = append(st.pinRules, rule.RuleID)
				}
				if _, seen := pinnedSeen[id]; !seen {
					pinnedOrder = append(pinnedOrder, id)
					pinnedSeen[id] = struct{}{}
				}
				remainingPins--
				rulePinned++
			}

		case types.RuleActionBoost:
			if rule.BoostValue == nil || *rule.BoostValue == 0 {
				continue
			}
			for _, id := range matched {
				st := ensureState(states, id)
				if st.blocked {
					continue
				}
				cand, ok := candidateMap[id]
				if !ok {
					cand = types.ScoredItem{ItemID: id, Score: 0}
					candidateMap[id] = cand
					order = append(order, id)
				}
				cand.Score += *rule.BoostValue
				candidateMap[id] = cand
				st.boostDelta += *rule.BoostValue
				st.boostRules = append(st.boostRules, BoostDetail{RuleID: rule.RuleID, Delta: *rule.BoostValue})
			}
		}
	}

	// Build final candidate list excluding blocked/pinned items.
	finalCands := make([]types.ScoredItem, 0, len(order))
	for _, id := range order {
		cand, ok := candidateMap[id]
		if !ok {
			continue
		}
		st := states[id]
		if st != nil {
			if st.blocked {
				continue
			}
			if st.pinned {
				continue
			}
		}
		finalCands = append(finalCands, cand)
	}
	result.Candidates = finalCands

	// Build pinned items payload.
	pinnedItems := make([]PinnedItem, 0, len(pinnedOrder))
	for _, id := range pinnedOrder {
		st := states[id]
		if st == nil || !st.pinned || st.blocked {
			continue
		}
		cand, ok := candidateMap[id]
		fromCandidates := ok
		pinnedItems = append(pinnedItems, PinnedItem{
			ItemID:         id,
			Score:          cand.Score,
			FromCandidates: fromCandidates,
			Rules:          append([]uuid.UUID(nil), st.pinRules...),
		})
	}
	result.Pinned = pinnedItems

	// Assemble item effects and reasons.
	for id, st := range states {
		effect := ItemEffect{}
		if st.blocked {
			effect.Blocked = true
			effect.BlockRules = append([]uuid.UUID(nil), st.blockRules...)
			result.ReasonTags[id] = appendReason(result.ReasonTags[id], reasonTokens("rule.block", st.blockRules)...)
		}
		if st.pinned {
			effect.Pinned = true
			effect.PinRules = append([]uuid.UUID(nil), st.pinRules...)
			result.ReasonTags[id] = appendReason(result.ReasonTags[id], reasonTokens("rule.pin", st.pinRules)...)
		}
		if st.boostDelta != 0 {
			effect.BoostDelta = st.boostDelta
			effect.BoostRules = append([]BoostDetail(nil), st.boostRules...)
			result.ReasonTags[id] = appendReason(result.ReasonTags[id], boostReasons(st.boostRules)...)
		}
		if effect.Blocked || effect.Pinned || effect.BoostDelta != 0 {
			result.ItemEffects[id] = effect
		}
	}

	result.Matches = matches
	result.EvaluatedRuleIDs = deduplicateUUIDs(evaluated)

	return result, nil
}

func (e *evaluator) prepareIndexes(itemTags map[string][]string) (map[string]map[string]struct{}, map[string][]string, map[string][]string) {
	toMatchers := func(prefixes []string) []string {
		if len(prefixes) == 0 {
			return nil
		}
		seen := make(map[string]struct{}, len(prefixes))
		out := make([]string, 0, len(prefixes))
		for _, raw := range prefixes {
			trimmed := strings.ToLower(strings.TrimSpace(raw))
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

	brandMatchers := toMatchers(e.brandTagPrefixes)
	categoryMatchers := toMatchers(e.categoryTagPrefixes)

	tagSets := make(map[string]map[string]struct{}, len(itemTags))
	brandValues := make(map[string][]string)
	categoryValues := make(map[string][]string)

	for itemID, tags := range itemTags {
		if len(tags) == 0 {
			continue
		}
		set := make(map[string]struct{}, len(tags))
		for _, tag := range tags {
			lower := strings.ToLower(strings.TrimSpace(tag))
			if lower == "" {
				continue
			}
			if val, ok := matchStructured(lower, brandMatchers); ok {
				appendUnique(brandValues, itemID, val)
				continue
			}
			if val, ok := matchStructured(lower, categoryMatchers); ok {
				appendUnique(categoryValues, itemID, val)
				continue
			}
			set[lower] = struct{}{}
		}
		if len(set) > 0 {
			tagSets[itemID] = set
		}
	}

	return tagSets, brandValues, categoryValues
}

func (e *evaluator) matchRule(
	rule types.Rule,
	candidateMap map[string]types.ScoredItem,
	tagSets map[string]map[string]struct{},
	brandValues map[string][]string,
	categoryValues map[string][]string,
) []string {
	switch rule.TargetType {
	case types.RuleTargetItem:
		if len(rule.ItemIDs) == 0 {
			return []string{}
		}
		ids := make([]string, 0, len(rule.ItemIDs))
		for _, id := range rule.ItemIDs {
			trimmed := strings.TrimSpace(id)
			if trimmed == "" {
				continue
			}
			ids = append(ids, trimmed)
		}
		return ids

	case types.RuleTargetTag:
		key := strings.ToLower(strings.TrimSpace(rule.TargetKey))
		if key == "" {
			return nil
		}
		return matchBySet(key, tagSets)

	case types.RuleTargetBrand:
		key := strings.ToLower(strings.TrimSpace(rule.TargetKey))
		if key == "" {
			return nil
		}
		return matchByList(key, brandValues)

	case types.RuleTargetCategory:
		key := strings.ToLower(strings.TrimSpace(rule.TargetKey))
		if key == "" {
			return nil
		}
		return matchByList(key, categoryValues)
	}
	return nil
}

func matchBySet(target string, sets map[string]map[string]struct{}) []string {
	out := make([]string, 0)
	for itemID, set := range sets {
		if _, ok := set[target]; ok {
			out = append(out, itemID)
		}
	}
	return out
}

func matchByList(target string, values map[string][]string) []string {
	out := make([]string, 0)
	for itemID, list := range values {
		for _, v := range list {
			if v == target {
				out = append(out, itemID)
				break
			}
		}
	}
	return out
}

func ensureState(states map[string]*itemState, id string) *itemState {
	st := states[id]
	if st == nil {
		st = &itemState{}
		states[id] = st
	}
	return st
}

func deduplicatePreserveOrder(ids []string) []string {
	if len(ids) <= 1 {
		return append([]string(nil), ids...)
	}
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func containsUUID(haystack []uuid.UUID, needle uuid.UUID) bool {
	for _, id := range haystack {
		if id == needle {
			return true
		}
	}
	return false
}

func deduplicateUUIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) <= 1 {
		return append([]uuid.UUID(nil), ids...)
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func appendReason(existing []string, additions ...string) []string {
	if len(additions) == 0 {
		if existing == nil {
			return nil
		}
		return append([]string(nil), existing...)
	}
	seen := make(map[string]struct{}, len(existing)+len(additions))
	out := make([]string, 0, len(existing)+len(additions))
	for _, token := range existing {
		if token == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	for _, token := range additions {
		if token == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	if len(out) == 0 {
		return nil
	}
	sort.Strings(out)
	return out
}

func reasonTokens(prefix string, ruleIDs []uuid.UUID) []string {
	if len(ruleIDs) == 0 {
		return nil
	}
	tokens := make([]string, 0, len(ruleIDs))
	for _, id := range ruleIDs {
		tokens = append(tokens, prefix+"["+id.String()+"]")
	}
	return tokens
}

func boostReasons(details []BoostDetail) []string {
	tokens := make([]string, 0, len(details))
	for _, detail := range details {
		tokens = append(tokens, formatBoostReason(detail))
	}
	return tokens
}

func formatBoostReason(detail BoostDetail) string {
	return fmt.Sprintf("rule.boost:%+.2f[%s]", detail.Delta, detail.RuleID.String())
}
