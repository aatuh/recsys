package algorithm

import (
	"context"
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"recsys/internal/rules"
	"recsys/internal/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// Default recent anchors window days if not configured.
const defaultRecentAnchorsWindowDays = 30

// Default co-visitation window days if not configured.
const defaultCovisWindowDays = 30

// Limit anchors for performance.
const maxRecentAnchors = 10

// Limit co-visitation neighbors for performance.
const maxCovisNeighbors = 200

// Limit embedding neighbors for performance.
const maxEmbNeighbors = 200

// Maximum stored sample IDs for policy diagnostics.
const maxPolicySampleIDs = 20

// Model versions.
const modelVersionPopularity = "popularity_v1"
const modelVersionBlend = "blend_v1"

// Engine handles recommendation algorithm logic.
type Engine struct {
	config       Config
	store        types.RecAlgoStore
	rulesManager *rules.Manager
}

// NewEngine creates a new recommendation engine.
func NewEngine(config Config, store types.RecAlgoStore, rulesManager *rules.Manager) *Engine {
	return &Engine{
		config:       config,
		store:        store,
		rulesManager: rulesManager,
	}
}

// Recommend generates recommendations using the blended scoring approach.
func (e *Engine) Recommend(
	ctx context.Context, req Request,
) (*Response, *TraceData, error) {
	// Set defaults.
	k := req.K
	if k <= 0 {
		k = 20
	}

	sourceMetrics := make(map[string]SourceMetric)

	// Get popularity candidates.
	popStart := time.Now()
	popCandidates, err := e.getPopularityCandidates(
		ctx, req.OrgID, req.Namespace, k, req.Constraints,
	)
	if err != nil {
		return nil, nil, err
	}
	sourceMetrics["popularity"] = SourceMetric{Count: len(popCandidates), Duration: time.Since(popStart)}

	existing := make(map[string]struct{}, len(popCandidates))
	maxPopScore := 0.0
	for _, c := range popCandidates {
		existing[c.ItemID] = struct{}{}
		if c.Score > maxPopScore {
			maxPopScore = c.Score
		}
	}

	if req.InjectAnchors && len(req.AnchorItemIDs) > 0 {
		if maxPopScore <= 0 {
			maxPopScore = 1.0
		}
		for _, anchor := range req.AnchorItemIDs {
			anchor = strings.TrimSpace(anchor)
			if anchor == "" {
				continue
			}
			if _, ok := existing[anchor]; ok {
				continue
			}
			popCandidates = append(popCandidates, types.ScoredItem{ItemID: anchor, Score: maxPopScore})
			existing[anchor] = struct{}{}
		}
	}

	collabScores := make(map[string]float64)
	contentScores := make(map[string]float64)
	sessionScores := make(map[string]float64)
	if req.UserID != "" {
		start := time.Now()
		_, scores, err := e.getCollaborativeCandidates(ctx, req, existing, k)
		if err != nil {
			return nil, nil, err
		}
		for id, score := range scores {
			collabScores[id] = score
		}
		sourceMetrics["collaborative"] = SourceMetric{Count: len(scores), Duration: time.Since(start)}
	}

	if req.UserID != "" {
		start := time.Now()
		_, scores, err := e.getContentBasedCandidates(ctx, req, existing, k)
		if err != nil {
			return nil, nil, err
		}
		for id, score := range scores {
			contentScores[id] = score
		}
		sourceMetrics["content"] = SourceMetric{Count: len(scores), Duration: time.Since(start)}
	}
	if req.UserID != "" {
		start := time.Now()
		_, scores, err := e.getSessionCandidates(ctx, req, existing, k)
		if err != nil {
			return nil, nil, err
		}
		for id, score := range scores {
			sessionScores[id] = score
		}
		sourceMetrics["session"] = SourceMetric{Count: len(scores), Duration: time.Since(start)}
	}

	popScores := make(map[string]float64, len(popCandidates))
	for _, cand := range popCandidates {
		popScores[cand.ItemID] = cand.Score
	}

	mergeStart := time.Now()
	merged := e.mergeCandidates(popCandidates, popScores, collabScores, contentScores, sessionScores, k)
	sourceMetrics["merged"] = SourceMetric{Count: len(merged), Duration: time.Since(mergeStart)}

	// Track policy enforcement stats for observability.
	policySummary := PolicySummary{TotalCandidates: len(merged)}

	// Apply exclusions.
	excludeStart := time.Now()
	candidates, err := e.applyExclusions(ctx, merged, req, &policySummary)
	if err != nil {
		return nil, nil, err
	}
	sourceMetrics["post_exclusion"] = SourceMetric{Count: len(candidates), Duration: time.Since(excludeStart)}

	// Get tags for all candidates.
	tags, err := e.getCandidateTags(ctx, candidates, req)
	if err != nil {
		return nil, nil, err
	}

	// Enforce positive constraints that require item metadata (e.g., tag whitelists).
	candidates, tags = e.applyConstraintFilters(candidates, tags, req, &policySummary)
	sourceMetrics["post_exclusion"] = SourceMetric{Count: len(candidates), Duration: time.Since(excludeStart)}

	// Align secondary score maps with filtered candidates.
	if len(collabScores) > 0 {
		collabScores = filterScoreMapByCandidates(collabScores, candidates)
	}
	if len(contentScores) > 0 {
		contentScores = filterScoreMapByCandidates(contentScores, candidates)
	}
	if len(sessionScores) > 0 {
		sessionScores = filterScoreMapByCandidates(sessionScores, candidates)
	}

	// Get blend weights.
	weights := e.getBlendWeights(req)

	// Gather co-visitation and embedding signals.
	candidateData, err := e.gatherSignals(ctx, candidates, req, weights)
	if err != nil {
		return nil, nil, err
	}
	candidateData.Tags = tags
	for id, score := range collabScores {
		if score > candidateData.EmbScores[id] {
			candidateData.EmbScores[id] = score
		}
		candidateData.UsedEmb[id] = true
		candidateData.Collaborative[id] = true
	}
	for id, score := range contentScores {
		if score > candidateData.EmbScores[id] {
			candidateData.EmbScores[id] = score
		}
		candidateData.UsedEmb[id] = true
		candidateData.ContentBased[id] = true
	}
	for id, score := range sessionScores {
		if score > candidateData.EmbScores[id] {
			candidateData.EmbScores[id] = score
		}
		candidateData.UsedEmb[id] = true
		candidateData.SessionBased[id] = true
	}

	// Apply blended scoring.
	e.applyBlendedScoring(candidateData, weights)

	// Apply personalization boost.
	e.applyPersonalizationBoost(ctx, candidateData, req)

	// Snapshot candidates before rule application to detect injected items.
	preRuleIDs := make(map[string]struct{}, len(candidateData.Candidates))
	for _, cand := range candidateData.Candidates {
		preRuleIDs[cand.ItemID] = struct{}{}
	}

	// Apply rule engine adjustments before MMR/caps.
	var ruleResult *rules.EvaluateResult
	if e.rulesManager != nil && e.config.RulesEnabled {
		ruleResult, err = e.applyRules(ctx, req, candidateData)
		if err != nil {
			return nil, nil, err
		}
	}
	if ruleResult != nil {
		policySummary.RulePinCount = len(ruleResult.Pinned)
		blockCount := 0
		boostCount := 0
		for _, effect := range ruleResult.ItemEffects {
			if effect.Blocked {
				blockCount++
			}
			if effect.BoostDelta != 0 {
				boostCount++
			}
		}
		policySummary.RuleBlockCount = blockCount
		policySummary.RuleBoostCount = boostCount
		// Track candidates injected by rules (e.g., manual boosts on new items).
		injected := 0
		for _, cand := range candidateData.Candidates {
			if _, ok := preRuleIDs[cand.ItemID]; !ok {
				injected++
			}
		}
		policySummary.RuleBoostInjected = injected
	} else {
		policySummary.RulePinCount = 0
		policySummary.RuleBlockCount = 0
		policySummary.RuleBoostCount = 0
		policySummary.RuleBoostInjected = 0
	}
	policySummary.AfterRules = len(candidateData.Candidates)

	// Ensure tags exist for any new candidates introduced by rules.
	if err := e.populateMissingTags(ctx, candidateData, req); err != nil {
		return nil, nil, err
	}

	// Determine model version.
	// Default requests (no explicit blend provided) are reported as
	// popularity_v1 to match API contract and tests, even if non-zero
	// config weights are used under the hood.
	modelVersion := modelVersionPopularity
	if req.Blend != nil {
		modelVersion = e.getModelVersion(weights)
	}
	kUsed := k
	candidatesPre := copyScoredItems(candidateData.Candidates)

	// Apply MMR and caps if needed.
	if e.shouldUseMMR() || e.shouldUseCaps() {
		reRanked, mmrInfo, capsInfo := mmrReRankInternal(
			candidateData.Candidates,
			candidateData.Tags,
			k,
			e.config.MMRLambda,
			e.config.BrandCap,
			e.config.CategoryCap,
			e.config.BrandTagPrefixes,
			e.config.CategoryTagPrefixes,
		)
		candidateData.Candidates = reRanked
		for id, info := range mmrInfo {
			candidateData.MMRInfo[id] = info
		}
		for id, caps := range capsInfo {
			candidateData.CapsInfo[id] = caps
		}
	}

	trace := &TraceData{
		K:                  kUsed,
		CandidatesPre:      candidatesPre,
		MMRInfo:            copyMMRInfo(candidateData.MMRInfo),
		CapsInfo:           copyCapsInfo(candidateData.CapsInfo),
		Anchors:            append([]string(nil), candidateData.Anchors...),
		Boosted:            copyBoolMap(candidateData.Boosted),
		IncludeReasons:     req.IncludeReasons,
		ExplainLevel:       req.ExplainLevel,
		ModelVersion:       modelVersion,
		SourceMetrics:      sourceMetrics,
		StarterProfile:     copyFloatMap(req.StarterProfile),
		StarterBlendWeight: req.StarterBlendWeight,
		RecentEventCount:   req.RecentEventCount,
	}
	reasonSink := make(map[string][]string)
	response := e.buildResponse(
		candidateData,
		kUsed,
		modelVersion,
		req.IncludeReasons,
		req.ExplainLevel,
		weights,
		reasonSink,
		ruleResult,
	)
	finalizePolicySummary(&policySummary, response, ruleResult)
	trace.Policy = &policySummary
	if req.InjectAnchors && len(req.AnchorItemIDs) > 0 && len(response.Items) > 0 {
		promoteAnchors(response, req.AnchorItemIDs, 3)
	}
	if len(trace.Anchors) == 0 && candidateData.AnchorsFetched {
		trace.Anchors = append(trace.Anchors, "(no_recent_activity)")
	}
	if len(reasonSink) > 0 {
		trace.Reasons = reasonSink
	} else {
		trace.Reasons = make(map[string][]string)
	}
	if ruleResult != nil {
		if len(ruleResult.Matches) > 0 {
			trace.RuleMatches = append([]rules.Match(nil), ruleResult.Matches...)
		}
		if len(ruleResult.ItemEffects) > 0 {
			effects := make(map[string]rules.ItemEffect, len(ruleResult.ItemEffects))
			for id, eff := range ruleResult.ItemEffects {
				effects[id] = eff
			}
			trace.RuleEffects = effects
		}
		if len(ruleResult.EvaluatedRuleIDs) > 0 {
			trace.RuleEvaluated = append([]uuid.UUID(nil), ruleResult.EvaluatedRuleIDs...)
		}
		if len(ruleResult.Pinned) > 0 {
			trace.RulePinned = append([]rules.PinnedItem(nil), ruleResult.Pinned...)
		}
	}
	return response, trace, nil
}

func (e *Engine) applyRules(
	ctx context.Context,
	req Request,
	data *CandidateData,
) (*rules.EvaluateResult, error) {
	if e.rulesManager == nil || !e.config.RulesEnabled {
		return nil, nil
	}
	surface := strings.TrimSpace(req.Surface)
	if surface == "" {
		surface = "default"
	}
	candidates := append([]types.ScoredItem(nil), data.Candidates...)
	itemTags := make(map[string][]string, len(data.Tags))
	for id, tags := range data.Tags {
		itemTags[id] = append([]string(nil), tags.Tags...)
	}

	evalReq := rules.EvaluateRequest{
		OrgID:               req.OrgID,
		Namespace:           req.Namespace,
		Surface:             surface,
		SegmentID:           req.SegmentID,
		Now:                 time.Now().UTC(),
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
		return nil, nil
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

// getPopularityCandidates fetches a popularity-based candidate pool.
// Uses configurable fanout: if <=0 -> k; if <k -> k.
func (e *Engine) getPopularityCandidates(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	k int,
	c *types.PopConstraints,
) ([]types.ScoredItem, error) {
	// Respect configured fanout if provided; otherwise default to k.
	fetchK := e.config.PopularityFanout
	if fetchK <= 0 || fetchK < k {
		fetchK = k
	}

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
) ([]types.ScoredItem, map[string]float64, error) {
	if req.UserID == "" {
		return nil, nil, nil
	}

	fetchK := e.config.PopularityFanout
	if fetchK <= 0 || fetchK < k {
		fetchK = k
	}

	exclude := make([]string, 0, len(existing))
	for id := range existing {
		exclude = append(exclude, id)
	}

	candidates, err := e.store.CollaborativeTopK(
		ctx,
		req.OrgID,
		req.Namespace,
		req.UserID,
		fetchK,
		exclude,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	scores := make(map[string]float64, len(candidates))
	filtered := make([]types.ScoredItem, 0, len(candidates))
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
		scores[cand.ItemID] = cand.Score
		filtered = append(filtered, types.ScoredItem{ItemID: cand.ItemID, Score: 0})
		existing[cand.ItemID] = struct{}{}
	}

	return filtered, scores, nil
}

func (e *Engine) getContentBasedCandidates(
	ctx context.Context,
	req Request,
	existing map[string]struct{},
	k int,
) ([]types.ScoredItem, map[string]float64, error) {
	if req.UserID == "" {
		return nil, nil, nil
	}

	profile, err := e.store.BuildUserTagProfile(
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
	sort.Slice(weights, func(i, j int) bool {
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

	fetchK := e.config.PopularityFanout
	if fetchK <= 0 || fetchK < k {
		fetchK = k
	}

	exclude := make([]string, 0, len(existing))
	for id := range existing {
		exclude = append(exclude, id)
	}

	candidates, err := e.store.ContentSimilarityTopK(
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
	filtered := make([]types.ScoredItem, 0, len(candidates))
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
		filtered = append(filtered, types.ScoredItem{ItemID: cand.ItemID, Score: 0})
		existing[cand.ItemID] = struct{}{}
	}

	return filtered, scores, nil
}

func (e *Engine) getSessionCandidates(
	ctx context.Context,
	req Request,
	existing map[string]struct{},
	k int,
) ([]types.ScoredItem, map[string]float64, error) {
	if req.UserID == "" {
		return nil, nil, nil
	}

	fetchK := e.config.PopularityFanout
	if fetchK <= 0 || fetchK < k {
		fetchK = k
	}

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

	candidates, err := e.store.SessionSequenceTopK(
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
	filtered := make([]types.ScoredItem, 0, len(candidates))
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
		filtered = append(filtered, types.ScoredItem{ItemID: cand.ItemID, Score: 0})
		existing[cand.ItemID] = struct{}{}
	}

	return filtered, scores, nil
}

// applyExclusions removes excluded items from candidates.
func (e *Engine) applyExclusions(
	ctx context.Context,
	candidates []types.ScoredItem,
	req Request,
	summary *PolicySummary,
) ([]types.ScoredItem, error) {
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
	exclude, recent, err = e.excludeRecentEventItems(ctx, req, exclude)
	if err != nil {
		return nil, err
	}

	// Filter candidates by excluding excluded items.
	filtered := make([]types.ScoredItem, 0, len(candidates))
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

	// Exclude purchased items in a time window.
	lookback := time.Duration(e.config.PurchasedWindowDays*24.0) * time.Hour
	since := time.Now().UTC().Add(-lookback)
	bought, err := e.store.ListUserEventsSince(
		ctx,
		req.OrgID,
		req.Namespace,
		req.UserID,
		since,
		e.config.ExcludeEventTypes,
	)
	if err != nil {
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
	ctx context.Context, candidates []types.ScoredItem, req Request,
) (map[string]types.ItemTags, error) {
	if len(candidates) == 0 {
		return make(map[string]types.ItemTags), nil
	}

	// Build list of candidate IDs.
	ids := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		ids = append(ids, candidate.ItemID)
	}

	// Fetch tags for all candidates.
	return e.store.ListItemsTags(ctx, req.OrgID, req.Namespace, ids)
}

func (e *Engine) applyConstraintFilters(
	candidates []types.ScoredItem,
	tags map[string]types.ItemTags,
	req Request,
	summary *PolicySummary,
) ([]types.ScoredItem, map[string]types.ItemTags) {
	if req.Constraints == nil {
		if summary != nil {
			summary.AfterConstraintFilters = len(candidates)
		}
		return candidates, tags
	}

	filtered := candidates[:0]
	prunedTags := tags
	removed := make([]string, 0)
	includeNormalized := make([]string, 0)

	if len(req.Constraints.IncludeTagsAny) > 0 {
		required := make(map[string]struct{}, len(req.Constraints.IncludeTagsAny))
		for _, tag := range req.Constraints.IncludeTagsAny {
			normalized := strings.ToLower(strings.TrimSpace(tag))
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
			pruned := make(map[string]types.ItemTags, len(tags))
			for _, cand := range candidates {
				itemTags, ok := tags[cand.ItemID]
				if !ok {
					removed = append(removed, cand.ItemID)
					continue
				}
				if hasAnyTag(itemTags.Tags, required) {
					filtered = append(filtered, cand)
					pruned[cand.ItemID] = itemTags
				} else {
					removed = append(removed, cand.ItemID)
				}
			}
			candidates = filtered
			prunedTags = pruned
		}
	}

	if summary != nil {
		if len(includeNormalized) > 0 {
			summary.ConstraintIncludeTags = append([]string(nil), includeNormalized...)
		} else {
			summary.ConstraintIncludeTags = nil
		}
		summary.ConstraintFilteredCount = len(removed)
		if len(removed) > 0 {
			if len(removed) > maxPolicySampleIDs {
				summary.ConstraintFilteredIDs = append([]string(nil), removed[:maxPolicySampleIDs]...)
			} else {
				summary.ConstraintFilteredIDs = append([]string(nil), removed...)
			}
			lookup := make(map[string]struct{}, len(removed))
			for _, id := range removed {
				lookup[id] = struct{}{}
			}
			summary.constraintFilteredLookup = lookup
		} else {
			summary.ConstraintFilteredIDs = nil
			summary.constraintFilteredLookup = nil
		}
		summary.AfterConstraintFilters = len(candidates)
	}

	return candidates, prunedTags
}

func hasAnyTag(candidateTags []string, required map[string]struct{}) bool {
	if len(candidateTags) == 0 || len(required) == 0 {
		return false
	}
	for _, tag := range candidateTags {
		if _, ok := required[strings.ToLower(strings.TrimSpace(tag))]; ok {
			return true
		}
	}
	return false
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

	if data.Tags == nil {
		data.Tags = make(map[string]types.ItemTags, len(tags))
	}
	for id, info := range tags {
		data.Tags[id] = info
	}
	return nil
}

func finalizePolicySummary(summary *PolicySummary, resp *Response, ruleResult *rules.EvaluateResult) {
	if summary == nil {
		return
	}
	if resp != nil {
		summary.FinalCount = len(resp.Items)
		if len(summary.constraintFilteredLookup) > 0 && len(resp.Items) > 0 {
			leaks := make([]string, 0)
			for _, item := range resp.Items {
				if _, ok := summary.constraintFilteredLookup[item.ItemID]; ok {
					leaks = append(leaks, item.ItemID)
				}
			}
			summary.ConstraintLeakCount = len(leaks)
			if len(leaks) > 0 {
				if len(leaks) > maxPolicySampleIDs {
					summary.ConstraintLeakIDs = append([]string(nil), leaks[:maxPolicySampleIDs]...)
				} else {
					summary.ConstraintLeakIDs = append([]string(nil), leaks...)
				}
			} else {
				summary.ConstraintLeakIDs = nil
			}
		} else {
			summary.ConstraintLeakCount = 0
			summary.ConstraintLeakIDs = nil
		}

		if ruleResult != nil && len(ruleResult.ItemEffects) > 0 {
			boostExposure := 0
			pinExposure := 0
			for _, item := range resp.Items {
				if eff, ok := ruleResult.ItemEffects[item.ItemID]; ok {
					if eff.BoostDelta != 0 {
						boostExposure++
					}
					if eff.Pinned {
						pinExposure++
					}
				}
			}
			summary.RuleBoostExposure = boostExposure
			summary.RulePinExposure = pinExposure
		} else {
			summary.RuleBoostExposure = 0
			summary.RulePinExposure = 0
		}
	}
	summary.constraintFilteredLookup = nil
}

func filterScoreMapByCandidates(scores map[string]float64, candidates []types.ScoredItem) map[string]float64 {
	if len(scores) == 0 {
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

// getBlendWeights returns the blend weights to use.
func (e *Engine) getBlendWeights(req Request) BlendWeights {
	weights := BlendWeights{
		Pop:  e.config.BlendAlpha,
		Cooc: e.config.BlendBeta,
		ALS:  e.config.BlendGamma,
	}

	// Override with request weights if provided.
	if req.Blend != nil {
		weights.Pop = math.Max(0, req.Blend.Pop)
		weights.Cooc = math.Max(0, req.Blend.Cooc)
		weights.ALS = math.Max(0, req.Blend.ALS)
	}

	// Safety: if all weights are 0, use popularity-only.
	if weights.Pop == 0 && weights.Cooc == 0 && weights.ALS == 0 {
		weights.Pop = 1
	}

	return weights
}

// gatherSignals collects recent anchors, co-visitation and embedding signals.
func (e *Engine) gatherSignals(
	ctx context.Context,
	candidates []types.ScoredItem,
	req Request,
	weights BlendWeights,
) (*CandidateData, error) {
	data := &CandidateData{
		Candidates:        candidates,
		CoocScores:        make(map[string]float64),
		EmbScores:         make(map[string]float64),
		UsedCooc:          make(map[string]bool),
		UsedEmb:           make(map[string]bool),
		Collaborative:     make(map[string]bool),
		ContentBased:      make(map[string]bool),
		SessionBased:      make(map[string]bool),
		Boosted:           make(map[string]bool),
		PopNorm:           make(map[string]float64),
		CoocNorm:          make(map[string]float64),
		EmbNorm:           make(map[string]float64),
		PopRaw:            make(map[string]float64),
		CoocRaw:           make(map[string]float64),
		EmbRaw:            make(map[string]float64),
		ProfileOverlap:    make(map[string]float64),
		ProfileMultiplier: make(map[string]float64),
		MMRInfo:           make(map[string]MMRExplain),
		CapsInfo:          make(map[string]CapsExplain),
	}

	// Build candidate set for quick lookup.
	candSet := make(map[string]struct{}, len(candidates))
	for _, c := range candidates {
		candSet[c.ItemID] = struct{}{}
	}

	// Only gather signals if we have a user.
	if req.UserID == "" {
		return data, nil
	}

	var anchors []string
	var err error
	if req.InjectAnchors && len(req.AnchorItemIDs) > 0 {
		anchors = append([]string(nil), req.AnchorItemIDs...)
		data.AnchorsFetched = true
	} else {
		// Recent anchors (views/purchases/etc.) for the user.
		anchors, err = e.getRecentAnchors(ctx, req)
		if err != nil {
			// Be resilient: still return a response with placeholders.
			data.AnchorsFetched = true
			data.Anchors = []string{"(no_recent_activity)"}
			return data, nil
		}
		data.AnchorsFetched = true
	}

	if len(anchors) == 0 {
		// Explicit placeholder when there is no history.
		data.Anchors = []string{"(no_recent_activity)"}
		return data, nil
	}

	// Deduplicate and keep order.
	seen := make(map[string]struct{}, len(anchors))
	for _, a := range anchors {
		if a == "" {
			continue
		}
		if _, ok := seen[a]; ok {
			continue
		}
		seen[a] = struct{}{}
		data.Anchors = append(data.Anchors, a)
	}

	// If we don’t use co-vis or embedding, we’re done (anchors are still useful
	// for explain).
	if weights.Cooc == 0 && weights.ALS == 0 {
		return data, nil
	}

	// Co-vis signals.
	if weights.Cooc > 0 {
		if err := e.gatherCoVisitationSignals(ctx, data, req, data.Anchors, candSet); err != nil {
			return nil, err
		}
	}

	// Embedding signals.
	if weights.ALS > 0 {
		if err := e.gatherEmbeddingSignals(ctx, data, req, data.Anchors, candSet); err != nil {
			return nil, err
		}
	}

	return data, nil
}

// getRecentAnchors gets recent items for the user to use as anchors.
func (e *Engine) getRecentAnchors(
	ctx context.Context, req Request,
) ([]string, error) {
	days := e.config.CoVisWindowDays
	if days <= 0 {
		days = defaultRecentAnchorsWindowDays
	}
	since := time.Now().UTC().Add(-time.Duration(days*24.0) * time.Hour)

	return e.store.ListUserRecentItemIDs(
		ctx,
		req.OrgID,
		req.Namespace,
		req.UserID,
		since,
		maxRecentAnchors,
	)
}

// gatherCoVisitationSignals collects co-visitation scores.
func (e *Engine) gatherCoVisitationSignals(
	ctx context.Context,
	data *CandidateData,
	req Request,
	anchors []string,
	candSet map[string]struct{},
) error {
	days := e.config.CoVisWindowDays
	if days <= 0 {
		days = defaultCovisWindowDays
	}
	since := time.Now().UTC().Add(-time.Duration(days*24.0) * time.Hour)

	for _, anchor := range anchors {
		// Gather co-visitation neighbors.
		neighbors, err := e.store.CooccurrenceTopKWithin(
			ctx,
			req.OrgID,
			req.Namespace,
			anchor,
			maxCovisNeighbors,
			since,
		)
		if err != nil {
			continue // Skip this anchor on error.
		}

		// Update co-visitation scores.
		for _, neighbor := range neighbors {
			if _, ok := candSet[neighbor.ItemID]; !ok {
				continue // Not in candidate set.
			}
			// Set score if higher than current score.
			if neighbor.Score > data.CoocScores[neighbor.ItemID] {
				data.CoocScores[neighbor.ItemID] = neighbor.Score
				data.UsedCooc[neighbor.ItemID] = true
			}
		}
	}

	return nil
}

// gatherEmbeddingSignals collects embedding similarity scores.
func (e *Engine) gatherEmbeddingSignals(
	ctx context.Context,
	data *CandidateData,
	req Request,
	anchors []string,
	candSet map[string]struct{},
) error {
	for _, anchor := range anchors {
		// Gather embedding neighbors.
		neighbors, err := e.store.SimilarByEmbeddingTopK(
			ctx,
			req.OrgID,
			req.Namespace,
			anchor,
			maxEmbNeighbors,
		)
		if err != nil {
			continue // Skip this anchor on error.
		}

		// Update embedding scores.
		for _, neighbor := range neighbors {
			if _, ok := candSet[neighbor.ItemID]; !ok {
				continue // Not in candidate set.
			}
			// Max aggregate: set score if higher than current score.
			if neighbor.Score > data.EmbScores[neighbor.ItemID] {
				data.EmbScores[neighbor.ItemID] = neighbor.Score
				data.UsedEmb[neighbor.ItemID] = true
			}
		}
	}

	return nil
}

// applyBlendedScoring applies the blended scoring formula.
func (e *Engine) applyBlendedScoring(
	data *CandidateData, weights BlendWeights,
) {
	maxPop, maxCooc, maxEmb := e.findMaxEmbeddingScores(data)
	e.applyBlendedScoringWithWeights(data, weights, maxPop, maxCooc, maxEmb)
}

// findMaxEmbeddingScores finds the max scores for normalization.
func (e *Engine) findMaxEmbeddingScores(
	data *CandidateData,
) (float64, float64, float64) {
	maxPop := 0.0
	maxCooc := 0.0
	maxEmb := 0.0
	for _, candidate := range data.Candidates {
		if candidate.Score > maxPop {
			maxPop = candidate.Score
		}
	}
	for _, score := range data.CoocScores {
		if score > maxCooc {
			maxCooc = score
		}
	}
	for _, score := range data.EmbScores {
		if score > maxEmb {
			maxEmb = score
		}
	}
	return maxPop, maxCooc, maxEmb
}

// applyBlendedScoringWithWeights applies blended scoring with weights.
func (e *Engine) applyBlendedScoringWithWeights(
	data *CandidateData,
	weights BlendWeights,
	maxPop float64,
	maxCooc float64,
	maxEmb float64,
) {
	for i := range data.Candidates {
		id := data.Candidates[i].ItemID

		popRaw := data.Candidates[i].Score
		popNorm := 0.0
		if maxPop > 0 {
			popNorm = popRaw / maxPop
		}

		coocRaw := data.CoocScores[id]
		coocNorm := 0.0
		if maxCooc > 0 {
			coocNorm = coocRaw / maxCooc
		}

		embRaw := data.EmbScores[id]
		embNorm := 0.0
		if maxEmb > 0 {
			embNorm = embRaw / maxEmb
		}

		blended := weights.Pop*popNorm +
			weights.Cooc*coocNorm +
			weights.ALS*embNorm

		data.PopNorm[id] = popNorm
		data.CoocNorm[id] = coocNorm
		data.EmbNorm[id] = embNorm
		data.PopRaw[id] = popRaw
		data.CoocRaw[id] = coocRaw
		data.EmbRaw[id] = embRaw

		data.Candidates[i].Score = blended
	}
}

// applyPersonalizationBoost applies personalization boost based on user
// profile.
func (e *Engine) applyPersonalizationBoost(
	ctx context.Context, data *CandidateData, req Request,
) {
	if req.UserID == "" || e.config.ProfileBoost <= 0 {
		return
	}

	// Build the user tag profile. Profile is a map of tag:weight where weights
	// sum to 1.
	profile, err := e.store.BuildUserTagProfile(
		ctx,
		req.OrgID,
		req.Namespace,
		req.UserID,
		e.config.ProfileWindowDays,
		maxInt(e.config.ProfileTopNTags, 1),
	)
	if err != nil {
		profile = nil
	}
	if len(req.StarterProfile) > 0 {
		if len(profile) == 0 {
			profile = copyFloatMap(req.StarterProfile)
		} else if req.StarterBlendWeight > 0 {
			profile = blendTagProfiles(profile, req.StarterProfile, req.StarterBlendWeight)
		}
	}
	if len(profile) == 0 {
		return
	}

	anchorCount := effectiveAnchorCount(data.Anchors)
	eventCount := req.RecentEventCount
	if eventCount <= 0 {
		eventCount = anchorCount
	} else if anchorCount > 0 && eventCount > anchorCount {
		eventCount = anchorCount
	}
	minEvents := e.config.ProfileMinEventsForBoost
	if minEvents < 0 {
		minEvents = 0
	}
	coldScale := clampFloat(e.config.ProfileColdStartMultiplier, 0, 1)

	for i := range data.Candidates {
		itemId := data.Candidates[i].ItemID
		tags := data.Tags[itemId]

		overlap := 0.0
		for _, tag := range tags.Tags {
			if weight, ok := profile[tag]; ok {
				overlap += weight
			}
		}

		if overlap > 0 {
			multiplier := 1.0 + e.config.ProfileBoost*overlap
			attenuation := 1.0
			if minEvents > 0 {
				if eventCount == 0 || eventCount < minEvents {
					attenuation = coldScale
				}
			}
			if attenuation < 1.0 {
				multiplier = 1.0 + (multiplier-1.0)*attenuation
			}
			data.Candidates[i].Score *= multiplier
			data.Boosted[itemId] = true
			data.ProfileOverlap[itemId] = overlap
			data.ProfileMultiplier[itemId] = multiplier
		}
	}
}

// getModelVersion determines the model version based on weights.
func (e *Engine) getModelVersion(weights BlendWeights) string {
	if weights.Cooc == 0 && weights.ALS == 0 {
		return modelVersionPopularity
	}
	return modelVersionBlend
}

// shouldUseMMR returns true if MMR should be applied.
func (e *Engine) shouldUseMMR() bool {
	return e.config.MMRLambda > 0
}

// shouldUseCaps returns true if caps should be applied.
func (e *Engine) shouldUseCaps() bool {
	return e.config.BrandCap > 0 || e.config.CategoryCap > 0
}

// buildResponse builds the final response.

func (e *Engine) buildResponse(
	data *CandidateData,
	k int,
	modelVersion string,
	includeReasons bool,
	explainLevel ExplainLevel,
	weights BlendWeights,
	reasonSink map[string][]string,
	ruleResult *rules.EvaluateResult,
) *Response {
	var pinned []rules.PinnedItem
	var ruleReasonTags map[string][]string
	if ruleResult != nil {
		pinned = ruleResult.Pinned
		ruleReasonTags = ruleResult.ReasonTags
	}

	estimated := len(data.Candidates) + len(pinned)
	if estimated > k {
		estimated = k
	}
	response := &Response{
		ModelVersion: modelVersion,
		Items:        make([]ScoredItem, 0, maxInt(estimated, 0)),
	}

	remaining := k

	appendReasons := func(itemID string, base []string) []string {
		combined := base
		if len(ruleReasonTags) > 0 {
			if extra, ok := ruleReasonTags[itemID]; ok && len(extra) > 0 {
				combined = append(combined, extra...)
			}
		}
		return deduplicateReasons(combined)
	}

	// Attach pinned items first.
	for _, pin := range pinned {
		if remaining <= 0 {
			break
		}
		reasonsFull := appendReasons(pin.ItemID, e.buildReasons(pin.ItemID, true, weights, data))
		if reasonSink != nil {
			reasonSink[pin.ItemID] = append([]string(nil), reasonsFull...)
		}
		var reasons []string
		if includeReasons {
			reasons = append([]string(nil), reasonsFull...)
		}
		var explain *ExplainBlock
		if explainLevel != ExplainLevelTags {
			explain = e.buildExplain(pin.ItemID, weights, data, explainLevel)
		}
		score := pin.Score
		response.Items = append(response.Items, ScoredItem{
			ItemID:  pin.ItemID,
			Score:   score,
			Reasons: reasons,
			Explain: explain,
		})
		remaining--
	}

	if remaining <= 0 {
		return response
	}

	// Sort remaining candidates by score.
	sortedCandidates := make([]types.ScoredItem, len(data.Candidates))
	copy(sortedCandidates, data.Candidates)
	sort.Slice(sortedCandidates, func(i, j int) bool {
		return sortedCandidates[i].Score > sortedCandidates[j].Score
	})

	for _, candidate := range sortedCandidates {
		if remaining <= 0 {
			break
		}
		reasonsFull := appendReasons(candidate.ItemID, e.buildReasons(candidate.ItemID, true, weights, data))
		if reasonSink != nil {
			reasonSink[candidate.ItemID] = append([]string(nil), reasonsFull...)
		}
		var reasons []string
		if includeReasons {
			reasons = append([]string(nil), reasonsFull...)
		}
		var explain *ExplainBlock
		if explainLevel != ExplainLevelTags {
			explain = e.buildExplain(candidate.ItemID, weights, data, explainLevel)
		}
		response.Items = append(response.Items, ScoredItem{
			ItemID:  candidate.ItemID,
			Score:   candidate.Score,
			Reasons: reasons,
			Explain: explain,
		})
		remaining--
	}

	return response
}

// buildReasons builds the reasons for a scored item.
func (e *Engine) buildReasons(
	itemID string,
	includeReasons bool,
	weights BlendWeights,
	data *CandidateData,
) []string {
	if !includeReasons {
		return nil
	}

	var reasons []string

	if weights.Pop > 0 {
		reasons = append(reasons, "recent_popularity")
	}
	if data.Collaborative[itemID] {
		reasons = append(reasons, "collaborative")
	}
	if data.ContentBased[itemID] {
		reasons = append(reasons, "content_similarity")
	}
	if data.SessionBased[itemID] {
		reasons = append(reasons, "session_sequence")
	}

	if data.UsedCooc[itemID] && weights.Cooc > 0 {
		reasons = append(reasons, "co_visitation")
	}

	if data.UsedEmb[itemID] && weights.ALS > 0 {
		reasons = append(reasons, "embedding_similarity")
	}

	if e.shouldUseMMR() || e.shouldUseCaps() {
		reasons = append(reasons, "diversity")
	}

	if data.Boosted[itemID] {
		reasons = append(reasons, "personalization")
	}

	return deduplicateReasons(reasons)
}

// buildExplain builds the structured explanation block for a scored item.
func (e *Engine) buildExplain(
	itemID string,
	weights BlendWeights,
	data *CandidateData,
	level ExplainLevel,
) *ExplainBlock {
	explain := &ExplainBlock{}

	// Always carry anchors (or a placeholder for non-tags explain levels).
	if len(data.Anchors) > 0 {
		explain.Anchors = append([]string(nil), data.Anchors...)
	} else if level != ExplainLevelTags {
		explain.Anchors = []string{"(no_recent_activity)"}
	}

	if weights.Pop > 0 || weights.Cooc > 0 || weights.ALS > 0 {
		blend := &BlendExplain{
			Alpha:    weights.Pop,
			Beta:     weights.Cooc,
			Gamma:    weights.ALS,
			PopNorm:  data.PopNorm[itemID],
			CoocNorm: data.CoocNorm[itemID],
			EmbNorm:  data.EmbNorm[itemID],
			Contributions: BlendContribution{
				Pop:  weights.Pop * data.PopNorm[itemID],
				Cooc: weights.Cooc * data.CoocNorm[itemID],
				Emb:  weights.ALS * data.EmbNorm[itemID],
			},
		}
		if level == ExplainLevelFull {
			blend.Raw = &BlendRaw{
				Pop:  data.PopRaw[itemID],
				Cooc: data.CoocRaw[itemID],
				Emb:  data.EmbRaw[itemID],
			}
		}
		explain.Blend = blend
	}

	if overlap, ok := data.ProfileOverlap[itemID]; ok {
		multiplier := data.ProfileMultiplier[itemID]
		pers := &PersonalizationExplain{
			Overlap:         overlap,
			BoostMultiplier: multiplier,
		}
		if level == ExplainLevelFull {
			pers.Raw = &PersonalizationExplainRaw{
				ProfileBoost: e.config.ProfileBoost,
			}
		}
		explain.Personalization = pers
	}

	if info, ok := data.MMRInfo[itemID]; ok {
		mmr := MMRExplain{
			Lambda:        info.Lambda,
			MaxSimilarity: info.MaxSimilarity,
			Penalty:       info.Penalty,
		}
		if level == ExplainLevelFull {
			mmr.Relevance = info.Relevance
			mmr.Rank = info.Rank
		}
		explain.MMR = &mmr
	} else if level == ExplainLevelFull && e.shouldUseMMR() {
		mmr := MMRExplain{
			Lambda: e.config.MMRLambda,
		}
		explain.MMR = &mmr
	}

	if caps, ok := data.CapsInfo[itemID]; ok {
		capsOut := &CapsExplain{}
		if caps.Brand != nil {
			usage := &CapUsage{Applied: caps.Brand.Applied}
			if level == ExplainLevelFull {
				if caps.Brand.Value != "" {
					usage.Value = caps.Brand.Value
				}
				if caps.Brand.Count != nil {
					count := *caps.Brand.Count
					usage.Count = &count
				}
				if caps.Brand.Limit != nil {
					limit := *caps.Brand.Limit
					usage.Limit = &limit
				}
			}
			capsOut.Brand = usage
		}
		if caps.Category != nil {
			usage := &CapUsage{Applied: caps.Category.Applied}
			if level == ExplainLevelFull {
				if caps.Category.Value != "" {
					usage.Value = caps.Category.Value
				}
				if caps.Category.Count != nil {
					count := *caps.Category.Count
					usage.Count = &count
				}
				if caps.Category.Limit != nil {
					limit := *caps.Category.Limit
					usage.Limit = &limit
				}
			}
			capsOut.Category = usage
		}
		if capsOut.Brand != nil || capsOut.Category != nil {
			explain.Caps = capsOut
		}
	} else if level == ExplainLevelFull && e.shouldUseCaps() {
		capsOut := &CapsExplain{}
		if e.config.BrandCap > 0 {
			limit := e.config.BrandCap
			capsOut.Brand = &CapUsage{Applied: false, Limit: &limit}
		}
		if e.config.CategoryCap > 0 {
			limit := e.config.CategoryCap
			capsOut.Category = &CapUsage{Applied: false, Limit: &limit}
		}
		if capsOut.Brand != nil || capsOut.Category != nil {
			explain.Caps = capsOut
		}
	}

	if explain.Blend == nil && explain.Personalization == nil && explain.MMR == nil && explain.Caps == nil && len(explain.Anchors) == 0 {
		return nil
	}

	return explain
}

func copyScoredItems(items []types.ScoredItem) []types.ScoredItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]types.ScoredItem, len(items))
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

func blendTagProfiles(primary, starter map[string]float64, starterWeight float64) map[string]float64 {
	weight := clampFloat(starterWeight, 0, 1)
	if weight == 0 {
		return copyFloatMap(primary)
	}
	if len(primary) == 0 {
		return copyFloatMap(starter)
	}

	blended := make(map[string]float64, len(primary)+len(starter))
	profileWeight := 1 - weight
	if profileWeight > 0 {
		for k, v := range primary {
			if v <= 0 {
				continue
			}
			blended[k] = v * profileWeight
		}
	}
	if weight > 0 {
		for k, v := range starter {
			if v <= 0 {
				continue
			}
			blended[k] += v * weight
		}
	}
	normalizeFloatMap(blended)
	return blended
}

func normalizeFloatMap(values map[string]float64) {
	if len(values) == 0 {
		return
	}
	sum := 0.0
	for _, v := range values {
		if v > 0 {
			sum += v
		}
	}
	if sum <= 0 {
		for k := range values {
			delete(values, k)
		}
		return
	}
	for k, v := range values {
		if v <= 0 {
			delete(values, k)
			continue
		}
		values[k] = v / sum
	}
}

func effectiveAnchorCount(anchors []string) int {
	if len(anchors) == 0 {
		return 0
	}
	if len(anchors) == 1 && anchors[0] == "(no_recent_activity)" {
		return 0
	}
	return len(anchors)
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

// deduplicateReasons deduplicates reasons.
func deduplicateReasons(reasons []string) []string {
	seen := make(map[string]struct{}, len(reasons))
	out := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		if reason == "" {
			continue
		}
		if _, ok := seen[reason]; ok {
			continue
		}
		seen[reason] = struct{}{}
		out = append(out, reason)
	}
	return out
}

func (e *Engine) mergeCandidates(
	pop []types.ScoredItem,
	popScores map[string]float64,
	collab map[string]float64,
	content map[string]float64,
	session map[string]float64,
	k int,
) []types.ScoredItem {
	maxKeep := len(pop)
	if k > maxKeep {
		maxKeep = k
	}
	if e.config.PopularityFanout > maxKeep {
		maxKeep = e.config.PopularityFanout
	}
	if maxKeep <= 0 {
		maxKeep = 1
	}

	pool := make([]types.ScoredItem, 0, maxKeep)
	used := make(map[string]struct{}, maxKeep)

	appendIfNew := func(itemID string, score float64) bool {
		if len(pool) >= maxKeep {
			return false
		}
		if itemID == "" {
			return false
		}
		if _, ok := used[itemID]; ok {
			return false
		}
		used[itemID] = struct{}{}
		pool = append(pool, types.ScoredItem{ItemID: itemID, Score: score})
		return true
	}

	for _, cand := range pop {
		if !appendIfNew(cand.ItemID, cand.Score) && len(pool) >= maxKeep {
			break
		}
	}

	reserveSlots := max(len(pop)/4, 20)
	if reserveSlots > len(pool) {
		reserveSlots = len(pool)
	}
	reserveBuffer := make([]types.ScoredItem, 0, reserveSlots)
	for i := 0; i < reserveSlots; i++ {
		lastIdx := len(pool) - 1
		if lastIdx < 0 {
			break
		}
		removed := pool[lastIdx]
		pool = pool[:lastIdx]
		delete(used, removed.ItemID)
		reserveBuffer = append(reserveBuffer, removed)
	}

	quotaOther := reserveSlots
	appendFromMap := func(src map[string]float64) {
		if quotaOther <= 0 {
			return
		}
		type pair struct {
			id    string
			score float64
		}
		items := make([]pair, 0, len(src))
		for id, score := range src {
			items = append(items, pair{id: id, score: score})
		}
		sort.Slice(items, func(i, j int) bool {
			if items[i].score == items[j].score {
				return items[i].id < items[j].id
			}
			return items[i].score > items[j].score
		})
		for _, it := range items {
			if quotaOther <= 0 {
				break
			}
			score := it.score
			if base, ok := popScores[it.id]; ok {
				score = base
			}
			if appendIfNew(it.id, score) {
				quotaOther--
			}
		}
	}

	appendFromMap(collab)
	appendFromMap(content)
	appendFromMap(session)

	for _, cand := range reserveBuffer {
		appendIfNew(cand.ItemID, cand.Score)
	}

	return pool
}
