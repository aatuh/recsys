package algorithm

import (
	"context"
	"strings"

	"github.com/aatuh/recsys-algo/rules"

	recmodel "github.com/aatuh/recsys-algo/model"
)

func (e *Engine) applyRules(
	ctx context.Context,
	req Request,
	data *CandidateData,
) (*rules.EvaluateResult, error) {
	noRuleResult := func() *rules.EvaluateResult {
		if data == nil {
			return &rules.EvaluateResult{}
		}
		return &rules.EvaluateResult{
			Candidates: append([]recmodel.ScoredItem(nil), data.Candidates...),
		}
	}
	if e.rulesManager == nil || !e.config.RulesEnabled {
		return noRuleResult(), nil
	}
	surface := strings.TrimSpace(req.Surface)
	if surface == "" {
		surface = "default"
	}
	candidates := append([]recmodel.ScoredItem(nil), data.Candidates...)
	itemTags := make(map[string][]string, len(data.Tags))
	for id, tags := range data.Tags {
		itemTags[id] = append([]string(nil), tags.Tags...)
	}

	evalReq := rules.EvaluateRequest{
		OrgID:               req.OrgID,
		Namespace:           req.Namespace,
		Surface:             surface,
		SegmentID:           req.SegmentID,
		Now:                 e.clock.Now(),
		Candidates:          candidates,
		ItemTags:            itemTags,
		BrandTagPrefixes:    e.config.BrandTagPrefixes,
		CategoryTagPrefixes: e.config.CategoryTagPrefixes,
	}

	result, err := e.rulesManager.Evaluate(ctx, evalReq)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return noRuleResult(), nil
	}
	data.Candidates = result.Candidates
	return result, nil
}

func promoteAnchors(resp *Response, anchors []string, maxPromote int) {
	if maxPromote <= 0 || resp == nil || len(resp.Items) == 0 || len(anchors) == 0 {
		return
	}
	anchorOrder := make([]string, 0, len(anchors))
	seenAnchors := make(map[string]struct{}, len(anchors))
	for _, anchor := range anchors {
		id := strings.TrimSpace(anchor)
		if id == "" {
			continue
		}
		if _, ok := seenAnchors[id]; ok {
			continue
		}
		anchorOrder = append(anchorOrder, id)
		seenAnchors[id] = struct{}{}
	}
	if len(anchorOrder) == 0 {
		return
	}

	type itemPos struct {
		idx int
		val ScoredItem
	}

	itemIndex := make(map[string]itemPos, len(resp.Items))
	for idx, item := range resp.Items {
		itemIndex[item.ItemID] = itemPos{idx: idx, val: item}
	}

	promoted := make([]ScoredItem, 0, maxPromote)
	promotedSet := make(map[string]struct{}, maxPromote)
	for _, anchor := range anchorOrder {
		if len(promoted) >= maxPromote {
			break
		}
		if pos, ok := itemIndex[anchor]; ok {
			promoted = append(promoted, pos.val)
			promotedSet[anchor] = struct{}{}
		}
	}
	if len(promoted) == 0 {
		return
	}

	rest := make([]ScoredItem, 0, len(resp.Items)-len(promoted))
	for _, item := range resp.Items {
		if _, ok := promotedSet[item.ItemID]; ok {
			continue
		}
		rest = append(rest, item)
	}
	resp.Items = append(promoted, rest...)
}

func promoteManualBoosts(resp *Response, ruleResult *rules.EvaluateResult) {
	if resp == nil || ruleResult == nil || len(resp.Items) == 0 {
		return
	}

	boostedSet := make(map[string]struct{})
	for id, effect := range ruleResult.ItemEffects {
		if effect.BoostDelta > 0 {
			boostedSet[id] = struct{}{}
		}
	}
	if len(boostedSet) == 0 {
		return
	}

	pinnedSet := make(map[string]struct{}, len(ruleResult.Pinned))
	for _, pin := range ruleResult.Pinned {
		pinnedSet[pin.ItemID] = struct{}{}
	}

	boosted := make([]ScoredItem, 0, len(boostedSet))
	for _, item := range resp.Items {
		if _, pinned := pinnedSet[item.ItemID]; pinned {
			continue
		}
		if _, ok := boostedSet[item.ItemID]; ok {
			boosted = append(boosted, item)
		}
	}
	if len(boosted) == 0 {
		return
	}

	pinnedItems := make([]ScoredItem, 0, len(pinnedSet))
	for _, item := range resp.Items {
		if _, pinned := pinnedSet[item.ItemID]; pinned {
			pinnedItems = append(pinnedItems, item)
		}
	}

	seen := make(map[string]struct{}, len(resp.Items))
	for _, item := range pinnedItems {
		seen[item.ItemID] = struct{}{}
	}
	for _, item := range boosted {
		seen[item.ItemID] = struct{}{}
	}

	rest := make([]ScoredItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		if _, skip := seen[item.ItemID]; skip {
			continue
		}
		rest = append(rest, item)
	}

	resp.Items = append(append(pinnedItems, boosted...), rest...)
}
