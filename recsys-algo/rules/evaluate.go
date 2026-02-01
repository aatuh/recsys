package rules

import (
	"context"
	"fmt"
	"sort"
	"strings"

	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"

	"github.com/google/uuid"
)

type evaluator struct {
	maxPinSlots         int
	brandTagPrefixes    []string
	categoryTagPrefixes []string
}

type itemState struct {
	blocked     bool
	blockRules  []uuid.UUID
	pinned      bool
	pinRules    []uuid.UUID
	boostDelta  float64
	boostRules  []BoostDetail
	pinPriority int
}

func (e *evaluator) apply(ctx context.Context, rules []Rule, req EvaluateRequest) (*EvaluateResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	result := &EvaluateResult{
		Candidates:       nil,
		Pinned:           nil,
		Matches:          nil,
		EvaluatedRuleIDs: nil,
		ItemEffects:      make(map[string]ItemEffect),
		ReasonTags:       make(map[string][]string),
	}
	overrideStats := make(map[uuid.UUID]*OverrideHit)

	if len(req.Candidates) == 0 {
		result.Candidates = []recmodel.ScoredItem{}
	}

	candidateMap := make(map[string]recmodel.ScoredItem, len(req.Candidates))
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
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		evaluated = append(evaluated, rule.RuleID)

		matched := e.matchRule(rule, tagSets, brandValues, categoryValues)
		if len(matched) == 0 {
			continue
		}

		// Deduplicate item IDs for the match payload while preserving order.
		matched = deduplicatePreserveOrder(matched)
		if stat := ensureOverrideHit(rule, overrideStats); stat != nil {
			stat.MatchedItems = appendUniqueStrings(stat.MatchedItems, matched)
		}
		matches = append(matches, Match{
			RuleID:           rule.RuleID,
			Action:           rule.Action,
			Target:           rule.TargetType,
			ItemIDs:          append([]string(nil), matched...),
			ManualOverrideID: rule.ManualOverrideID,
		})

		switch rule.Action {
		case RuleActionBlock:
			for _, id := range matched {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
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
				if stat := overrideStats[rule.RuleID]; stat != nil {
					stat.BlockedItems = appendUniqueStrings(stat.BlockedItems, []string{id})
				}
			}

		case RuleActionPin:
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
				if err := ctx.Err(); err != nil {
					return nil, err
				}
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
				if stat := overrideStats[rule.RuleID]; stat != nil {
					stat.PinnedItems = appendUniqueStrings(stat.PinnedItems, []string{id})
				}
				if _, seen := pinnedSeen[id]; !seen {
					pinnedOrder = append(pinnedOrder, id)
					pinnedSeen[id] = struct{}{}
				}
				remainingPins--
				rulePinned++
			}

		case RuleActionBoost:
			if rule.BoostValue == nil || *rule.BoostValue == 0 {
				continue
			}
			for _, id := range matched {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				st := ensureState(states, id)
				if st.blocked {
					continue
				}
				cand, ok := candidateMap[id]
				if !ok {
					cand = recmodel.ScoredItem{ItemID: id, Score: 0}
					candidateMap[id] = cand
					order = append(order, id)
				}
				current := cand.Score
				delta := current * *rule.BoostValue
				if delta == 0 {
					delta = *rule.BoostValue
				}
				cand.Score += delta
				candidateMap[id] = cand
				st.boostDelta += delta
				st.boostRules = append(st.boostRules, BoostDetail{RuleID: rule.RuleID, Delta: delta})
				if stat := overrideStats[rule.RuleID]; stat != nil {
					stat.BoostedItems = appendUniqueStrings(stat.BoostedItems, []string{id})
				}
			}
		}
	}

	// Build final candidate list excluding blocked/pinned items.
	finalCands := make([]recmodel.ScoredItem, 0, len(order))
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
	if len(overrideStats) > 0 {
		overrideList := make([]OverrideHit, 0, len(overrideStats))
		for _, hit := range overrideStats {
			overrideList = append(overrideList, *hit)
		}
		sort.Slice(overrideList, func(i, j int) bool {
			return overrideList[i].OverrideID.String() < overrideList[j].OverrideID.String()
		})
		result.OverrideHits = overrideList
		index := make(map[uuid.UUID]*OverrideHit, len(overrideList))
		for i := range overrideList {
			h := &result.OverrideHits[i]
			index[h.RuleID] = h
		}
		result.overrideIndex = index
	}

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
			lower := recmodel.NormalizeTag(tag)
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
	rule Rule,
	tagSets map[string]map[string]struct{},
	brandValues map[string][]string,
	categoryValues map[string][]string,
) []string {
	switch rule.TargetType {
	case RuleTargetItem:
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

	case RuleTargetTag:
		key := recmodel.NormalizeTag(rule.TargetKey)
		if key == "" {
			return nil
		}
		return matchBySet(key, tagSets)

	case RuleTargetBrand:
		key := recmodel.NormalizeTag(rule.TargetKey)
		if key == "" {
			return nil
		}
		return matchByList(key, brandValues)

	case RuleTargetCategory:
		key := recmodel.NormalizeTag(rule.TargetKey)
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

func ensureOverrideHit(rule Rule, stats map[uuid.UUID]*OverrideHit) *OverrideHit {
	if rule.ManualOverrideID == nil {
		return nil
	}
	if hit, ok := stats[rule.RuleID]; ok {
		return hit
	}
	hit := &OverrideHit{
		OverrideID: *rule.ManualOverrideID,
		RuleID:     rule.RuleID,
		Action:     rule.Action,
	}
	stats[rule.RuleID] = hit
	return hit
}

func appendUniqueStrings(dst []string, values []string) []string {
	if len(values) == 0 {
		return dst
	}
	seen := make(map[string]struct{}, len(dst))
	for _, v := range dst {
		seen[v] = struct{}{}
	}
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		dst = append(dst, v)
	}
	return dst
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
