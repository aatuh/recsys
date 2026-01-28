package algorithm

import (
	"context"
	"sort"

	recmodel "github.com/aatuh/recsys-algo/model"

	"github.com/google/uuid"
)

// getPopularityCandidates fetches a popularity-based candidate pool.
// Uses configurable fanout: if <=0 -> k; if <k -> k.
func (e *Engine) getPopularityCandidates(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	k int,
	c *recmodel.PopConstraints,
) ([]recmodel.ScoredItem, error) {
	// Respect configured fanout if provided; otherwise default to k.
	fetchK := e.fanoutFor(k)

	// Fetch popularity candidates.
	cands, err := e.store.PopularityTopK(
		ctx, orgID, ns, e.config.HalfLifeDays, fetchK, c,
	)
	if err != nil {
		return nil, err
	}

	return cands, nil
}

// getCollaborativeCandidates fetches ALS-based candidates for the user.
func (e *Engine) getCollaborativeCandidates(
	ctx context.Context,
	req Request,
	existing map[string]struct{},
	k int,
) ([]recmodel.ScoredItem, map[string]float64, error) {
	if req.UserID == "" {
		return nil, nil, nil
	}
	store, ok := e.store.(recmodel.CollaborativeStore)
	if !ok {
		return nil, nil, recmodel.ErrFeatureUnavailable
	}

	fetchK := e.fanoutFor(k)

	exclude := make([]string, 0, len(existing))
	for id := range existing {
		exclude = append(exclude, id)
	}

	candidates, err := store.CollaborativeTopK(
		ctx,
		req.OrgID,
		req.Namespace,
		req.UserID,
		fetchK,
		exclude,
	)
	if err != nil {
		return nil, nil, err
	}

	scores := make(map[string]float64, len(candidates))
	filtered := make([]recmodel.ScoredItem, 0, len(candidates))
	for _, cand := range candidates {
		if cand.ItemID == "" {
			continue
		}
		if cand.Score <= 0 {
			continue
		}
		if _, ok := existing[cand.ItemID]; ok {
			continue
		}
		scores[cand.ItemID] = cand.Score
		filtered = append(filtered, recmodel.ScoredItem{ItemID: cand.ItemID, Score: 0})
		existing[cand.ItemID] = struct{}{}
	}

	return filtered, scores, nil
}

func (e *Engine) getContentBasedCandidates(
	ctx context.Context,
	req Request,
	existing map[string]struct{},
	k int,
) ([]recmodel.ScoredItem, map[string]float64, error) {
	if req.UserID == "" {
		return nil, nil, nil
	}
	profileStore, ok := e.store.(recmodel.ProfileStore)
	if !ok {
		return nil, nil, recmodel.ErrFeatureUnavailable
	}
	contentStore, ok := e.store.(recmodel.ContentStore)
	if !ok {
		return nil, nil, recmodel.ErrFeatureUnavailable
	}

	profile, err := profileStore.BuildUserTagProfile(
		ctx,
		req.OrgID,
		req.Namespace,
		req.UserID,
		e.config.ProfileWindowDays,
		e.config.ProfileTopNTags,
	)
	if err != nil {
		return nil, nil, err
	}
	profile = normalizeTagWeights(profile)
	if len(req.StarterProfile) > 0 {
		if len(profile) == 0 {
			profile = copyFloatMap(req.StarterProfile)
		} else if req.StarterBlendWeight > 0 {
			profile = blendTagProfiles(profile, req.StarterProfile, req.StarterBlendWeight)
		}
	}
	if len(profile) == 0 {
		return nil, nil, nil
	}

	type tagWeight struct {
		tag    string
		weight float64
	}
	weights := make([]tagWeight, 0, len(profile))
	for tag, weight := range profile {
		if weight <= 0 {
			continue
		}
		weights = append(weights, tagWeight{tag: tag, weight: weight})
	}
	if len(weights) == 0 {
		return nil, nil, nil
	}
	sort.SliceStable(weights, func(i, j int) bool {
		if weights[i].weight == weights[j].weight {
			return weights[i].tag < weights[j].tag
		}
		return weights[i].weight > weights[j].weight
	})
	limit := len(weights)
	if e.config.ProfileTopNTags > 0 && limit > e.config.ProfileTopNTags {
		limit = e.config.ProfileTopNTags
	}
	tags := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		tags = append(tags, weights[i].tag)
	}

	if len(tags) == 0 {
		return nil, nil, nil
	}

	fetchK := e.fanoutFor(k)

	exclude := make([]string, 0, len(existing))
	for id := range existing {
		exclude = append(exclude, id)
	}

	candidates, err := contentStore.ContentSimilarityTopK(
		ctx,
		req.OrgID,
		req.Namespace,
		tags,
		fetchK,
		exclude,
	)
	if err != nil {
		return nil, nil, err
	}

	scores := make(map[string]float64, len(candidates))
	filtered := make([]recmodel.ScoredItem, 0, len(candidates))
	for _, cand := range candidates {
		if cand.ItemID == "" {
			continue
		}
		if cand.Score <= 0 {
			continue
		}
		if _, ok := existing[cand.ItemID]; ok {
			continue
		}
		scores[cand.ItemID] = cand.Score
		filtered = append(filtered, recmodel.ScoredItem{ItemID: cand.ItemID, Score: 0})
		existing[cand.ItemID] = struct{}{}
	}

	return filtered, scores, nil
}

func (e *Engine) getSessionCandidates(
	ctx context.Context,
	req Request,
	existing map[string]struct{},
	k int,
) ([]recmodel.ScoredItem, map[string]float64, error) {
	if req.UserID == "" {
		return nil, nil, nil
	}
	store, ok := e.store.(recmodel.SessionStore)
	if !ok {
		return nil, nil, recmodel.ErrFeatureUnavailable
	}

	fetchK := e.fanoutFor(k)

	exclude := make([]string, 0, len(existing))
	for id := range existing {
		exclude = append(exclude, id)
	}

	lookback := e.config.SessionLookbackEvents
	if lookback <= 0 {
		lookback = maxRecentAnchors
	}
	horizonMinutes := e.config.SessionLookaheadMinutes
	if horizonMinutes <= 0 {
		horizonMinutes = 30
	}

	candidates, err := store.SessionSequenceTopK(
		ctx,
		req.OrgID,
		req.Namespace,
		req.UserID,
		lookback,
		horizonMinutes,
		exclude,
		fetchK,
	)
	if err != nil {
		return nil, nil, err
	}

	scores := make(map[string]float64, len(candidates))
	filtered := make([]recmodel.ScoredItem, 0, len(candidates))
	for _, cand := range candidates {
		if cand.ItemID == "" {
			continue
		}
		if cand.Score <= 0 {
			continue
		}
		if _, ok := existing[cand.ItemID]; ok {
			continue
		}
		scores[cand.ItemID] = cand.Score
		filtered = append(filtered, recmodel.ScoredItem{ItemID: cand.ItemID, Score: 0})
		existing[cand.ItemID] = struct{}{}
	}

	return filtered, scores, nil
}

func filterScoreMapByCandidates(scores map[string]float64, candidates []recmodel.ScoredItem) map[string]float64 {
	if len(scores) == 0 || len(candidates) == 0 {
		return scores
	}
	filtered := make(map[string]float64, len(scores))
	for _, cand := range candidates {
		if score, ok := scores[cand.ItemID]; ok {
			filtered[cand.ItemID] = score
		}
	}
	return filtered
}

func (e *Engine) mergeCandidates(
	pop []recmodel.ScoredItem,
	collab map[string]float64,
	content map[string]float64,
	session map[string]float64,
	k int,
) ([]recmodel.ScoredItem, map[string]SourceSet) {
	maxKeep := len(pop)
	if k > maxKeep {
		maxKeep = k
	}
	if fanout := e.fanoutFor(k); fanout > maxKeep {
		maxKeep = fanout
	}
	if maxKeep <= 0 {
		maxKeep = 1
	}

	pool := make([]recmodel.ScoredItem, 0, maxKeep)
	used := make(map[string]struct{}, maxKeep)
	sources := make(map[string]SourceSet, maxKeep)

	appendIfNew := func(itemID string, score float64, limit int) bool {
		if len(pool) >= limit {
			return false
		}
		if itemID == "" {
			return false
		}
		if _, ok := used[itemID]; ok {
			return false
		}
		used[itemID] = struct{}{}
		pool = append(pool, recmodel.ScoredItem{ItemID: itemID, Score: score})
		return true
	}

	reserveSlots := 0
	if len(collab)+len(content)+len(session) > 0 {
		reserveSlots = len(pop) / 4
		if reserveSlots < 1 {
			reserveSlots = 1
		}
		maxReserve := minInt(20, len(pop))
		if reserveSlots > maxReserve {
			reserveSlots = maxReserve
		}
	}
	popLimit := maxKeep
	if reserveSlots > 0 && maxKeep > reserveSlots {
		popLimit = maxKeep - reserveSlots
	}

	popIndex := 0
	for ; popIndex < len(pop); popIndex++ {
		if len(pool) >= popLimit {
			break
		}
		cand := pop[popIndex]
		if !appendIfNew(cand.ItemID, cand.Score, maxKeep) {
			continue
		}
		addSource(sources, cand.ItemID, SignalPop)
	}

	for id, score := range collab {
		if !appendIfNew(id, score, maxKeep) {
			continue
		}
		addSource(sources, id, SignalCollaborative)
	}

	for id, score := range content {
		if !appendIfNew(id, score, maxKeep) {
			continue
		}
		addSource(sources, id, SignalContent)
	}

	for id, score := range session {
		if !appendIfNew(id, score, maxKeep) {
			continue
		}
		addSource(sources, id, SignalSession)
	}

	for ; popIndex < len(pop) && len(pool) < maxKeep; popIndex++ {
		cand := pop[popIndex]
		if !appendIfNew(cand.ItemID, cand.Score, maxKeep) {
			continue
		}
		addSource(sources, cand.ItemID, SignalPop)
	}

	return pool, sources
}

func copySourceSet(src SourceSet) SourceSet {
	if src == nil {
		return nil
	}
	copySet := make(SourceSet, len(src))
	for k := range src {
		copySet[k] = struct{}{}
	}
	return copySet
}

func filterSourcesByCandidates(
	sources map[string]SourceSet,
	candidates []recmodel.ScoredItem,
) map[string]SourceSet {
	if len(sources) == 0 || len(candidates) == 0 {
		return sources
	}
	filtered := make(map[string]SourceSet, len(candidates))
	for _, cand := range candidates {
		if set, ok := sources[cand.ItemID]; ok {
			filtered[cand.ItemID] = copySourceSet(set)
		}
	}
	return filtered
}
