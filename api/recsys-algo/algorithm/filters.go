package algorithm

import (
	"context"
	"errors"
	"strings"
	"time"

	recmodel "github.com/aatuh/recsys-algo/model"
)

// applyExclusions removes excluded items from candidates.
func (e *Engine) applyExclusions(
	ctx context.Context,
	candidates []recmodel.ScoredItem,
	req Request,
	summary *PolicySummary,
) ([]recmodel.ScoredItem, error) {
	exclude := make(map[string]struct{})
	explicit := make(map[string]struct{})

	// Add constraint exclusions.
	if req.Constraints != nil {
		for _, id := range req.Constraints.ExcludeItemIDs {
			trimmed := strings.TrimSpace(id)
			if trimmed == "" {
				continue
			}
			exclude[trimmed] = struct{}{}
			explicit[trimmed] = struct{}{}
		}
	}

	// Add exclusions from configured user events if enabled.
	var recent map[string]struct{}
	var err error
	_, recent, err = e.excludeRecentEventItems(ctx, req, exclude)
	if err != nil {
		return nil, err
	}

	// Filter candidates by excluding excluded items.
	filtered := make([]recmodel.ScoredItem, 0, len(candidates))
	for _, candidate := range candidates {
		_, skipExplicit := explicit[candidate.ItemID]
		_, skipRecent := recent[candidate.ItemID]
		if skipExplicit || skipRecent {
			if summary != nil {
				if skipExplicit {
					summary.ExplicitExcludeHits++
				} else if skipRecent {
					summary.RecentEventExcludeHits++
				}
			}
			continue
		}
		filtered = append(filtered, candidate)
	}

	if summary != nil {
		summary.AfterExclusions = len(filtered)
	}

	return filtered, nil
}

// excludeRecentEventItems excludes items linked to configured user events.
func (e *Engine) excludeRecentEventItems(
	ctx context.Context,
	req Request,
	exclude map[string]struct{},
) (map[string]struct{}, map[string]struct{}, error) {
	if !e.config.RuleExcludeEvents || req.UserID == "" {
		return exclude, make(map[string]struct{}), nil
	}
	eventStore, ok := e.store.(recmodel.EventStore)
	if !ok {
		return exclude, make(map[string]struct{}), nil
	}

	// Exclude purchased items in a time window.
	lookback := time.Duration(e.config.PurchasedWindowDays*24.0) * time.Hour
	since := e.clock.Now().Add(-lookback)
	bought, err := eventStore.ListUserEventsSince(
		ctx,
		req.OrgID,
		req.Namespace,
		req.UserID,
		since,
		e.config.ExcludeEventTypes,
	)
	if err != nil {
		if errors.Is(err, recmodel.ErrFeatureUnavailable) {
			return exclude, make(map[string]struct{}), nil
		}
		return nil, nil, err
	}

	// Add purchased items to exclude.
	recent := make(map[string]struct{}, len(bought))
	for _, id := range bought {
		exclude[id] = struct{}{}
		recent[id] = struct{}{}
	}

	return exclude, recent, nil
}

// getCandidateTags fetches tags for all candidates.
func (e *Engine) getCandidateTags(
	ctx context.Context, candidates []recmodel.ScoredItem, req Request,
) (map[string]recmodel.ItemTags, error) {
	if len(candidates) == 0 {
		return make(map[string]recmodel.ItemTags), nil
	}

	// Build list of candidate IDs.
	ids := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		ids = append(ids, candidate.ItemID)
	}

	// Fetch tags for all candidates.
	tags, err := e.store.ListItemsTags(ctx, req.OrgID, req.Namespace, ids)
	if err != nil {
		return nil, err
	}
	return normalizeItemTags(tags), nil
}

func (e *Engine) applyConstraintFilters(
	candidates []recmodel.ScoredItem,
	tags map[string]recmodel.ItemTags,
	req Request,
	summary *PolicySummary,
) ([]recmodel.ScoredItem, map[string]recmodel.ItemTags) {
	if req.Constraints == nil {
		if summary != nil {
			summary.AfterConstraintFilters = len(candidates)
		}
		return candidates, tags
	}

	working := candidates
	prunedTags := tags
	removed := make([]string, 0)
	var removedReasons map[string]string
	if summary != nil {
		removedReasons = make(map[string]string)
	}
	includeNormalized := make([]string, 0)

	recordRemoval := func(id, reason string) {
		if id == "" {
			return
		}
		removed = append(removed, id)
		if removedReasons != nil {
			if reason == "" {
				reason = constraintReasonUnknown
			}
			removedReasons[id] = reason
		}
	}

	if len(req.Constraints.IncludeTagsAny) > 0 {
		required := make(map[string]struct{}, len(req.Constraints.IncludeTagsAny))
		for _, tag := range req.Constraints.IncludeTagsAny {
			normalized := recmodel.NormalizeTag(tag)
			if normalized == "" {
				continue
			}
			if _, exists := required[normalized]; !exists {
				includeNormalized = append(includeNormalized, normalized)
			}
			required[normalized] = struct{}{}
		}
		if len(required) > 0 {
			// rebuild candidate list to only include items whose tags overlap required set.
			pruned := make(map[string]recmodel.ItemTags, len(prunedTags))
			next := working[:0]
			for _, cand := range working {
				itemTags, ok := prunedTags[cand.ItemID]
				if !ok {
					recordRemoval(cand.ItemID, constraintReasonMissingTags)
					continue
				}
				if hasAnyTag(itemTags.Tags, required) {
					next = append(next, cand)
					pruned[cand.ItemID] = itemTags
				} else {
					recordRemoval(cand.ItemID, constraintReasonInclude)
				}
			}
			working = next
			prunedTags = pruned
		}
	}

	// Apply min/max price constraints.
	if req.Constraints.MinPrice != nil || req.Constraints.MaxPrice != nil {
		pruned := make(map[string]recmodel.ItemTags, len(prunedTags))
		next := working[:0]
		for _, cand := range working {
			itemTags, ok := prunedTags[cand.ItemID]
			if !ok {
				recordRemoval(cand.ItemID, constraintReasonMissingTags)
				continue
			}
			if violates, reason := priceViolationReason(itemTags.Price, req.Constraints.MinPrice, req.Constraints.MaxPrice); violates {
				recordRemoval(cand.ItemID, reason)
				continue
			}
			next = append(next, cand)
			pruned[cand.ItemID] = itemTags
		}
		working = next
		prunedTags = pruned
	}

	// Apply created_after constraint.
	if req.Constraints.CreatedAfter != nil {
		pruned := make(map[string]recmodel.ItemTags, len(prunedTags))
		next := working[:0]
		for _, cand := range working {
			itemTags, ok := prunedTags[cand.ItemID]
			if !ok {
				recordRemoval(cand.ItemID, constraintReasonMissingTags)
				continue
			}
			if itemTags.CreatedAt.IsZero() || itemTags.CreatedAt.Before(*req.Constraints.CreatedAfter) {
				recordRemoval(cand.ItemID, constraintReasonCreated)
				continue
			}
			next = append(next, cand)
			pruned[cand.ItemID] = itemTags
		}
		working = next
		prunedTags = pruned
	}

	if summary != nil {
		if len(removed) > 0 {
			summary.ConstraintFilteredCount = len(removed)
			if len(removed) > maxPolicySampleIDs {
				summary.ConstraintFilteredIDs = append([]string(nil), removed[:maxPolicySampleIDs]...)
			} else {
				summary.ConstraintFilteredIDs = append([]string(nil), removed...)
			}
			summary.ConstraintIncludeTags = append([]string(nil), includeNormalized...)

			lookup := make(map[string]struct{}, len(removed))
			for _, id := range removed {
				lookup[id] = struct{}{}
			}
			summary.constraintFilteredLookup = lookup

			if len(removedReasons) > 0 {
				reasons := make(map[string]string, len(removedReasons))
				for id, reason := range removedReasons {
					reasons[id] = reason
				}
				summary.constraintFilteredReasons = reasons
			} else {
				summary.constraintFilteredReasons = nil
			}
		} else {
			summary.ConstraintFilteredCount = 0
			summary.ConstraintFilteredIDs = nil
			summary.ConstraintIncludeTags = nil
			summary.constraintFilteredLookup = nil
			summary.constraintFilteredReasons = nil
		}
		summary.AfterConstraintFilters = len(working)
	}

	return working, prunedTags
}

func hasAnyTag(candidateTags []string, required map[string]struct{}) bool {
	if len(candidateTags) == 0 || len(required) == 0 {
		return false
	}
	for _, tag := range candidateTags {
		if _, ok := required[recmodel.NormalizeTag(tag)]; ok {
			return true
		}
	}
	return false
}

func priceViolationReason(price *float64, minPrice, maxPrice *float64) (bool, string) {
	if price == nil {
		if minPrice != nil || maxPrice != nil {
			return true, constraintReasonPriceMiss
		}
		return false, ""
	}
	if minPrice != nil && *price < *minPrice {
		return true, constraintReasonPriceMin
	}
	if maxPrice != nil && *price > *maxPrice {
		return true, constraintReasonPriceMax
	}
	return false, ""
}

func (e *Engine) populateMissingTags(
	ctx context.Context,
	data *CandidateData,
	req Request,
) error {
	if data == nil {
		return nil
	}
	if e.store == nil {
		return nil
	}

	missing := make([]string, 0)
	for _, cand := range data.Candidates {
		if _, ok := data.Tags[cand.ItemID]; !ok {
			missing = append(missing, cand.ItemID)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	tags, err := e.store.ListItemsTags(ctx, req.OrgID, req.Namespace, missing)
	if err != nil {
		return err
	}
	tags = normalizeItemTags(tags)

	if data.Tags == nil {
		data.Tags = make(map[string]recmodel.ItemTags, len(tags))
	}
	for id, info := range tags {
		data.Tags[id] = info
	}
	return nil
}
