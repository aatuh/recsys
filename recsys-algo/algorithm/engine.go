package algorithm

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/api/recsys-algo/rules"

	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"

	"github.com/google/uuid"
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

const (
	constraintReasonInclude     = "include_tags"
	constraintReasonMissingTags = "missing_tags"
	constraintReasonPriceMin    = "price_below_min"
	constraintReasonPriceMax    = "price_above_max"
	constraintReasonPriceMiss   = "price_missing"
	constraintReasonCreated     = "stale_item"
	constraintReasonUnknown     = "unknown"
)

// Engine handles recommendation algorithm logic.
type Engine struct {
	config         Config
	store          recmodel.EngineStore
	rulesManager   *rules.Manager
	clock          Clock
	signalObserver SignalObserver
}

// NewEngine creates a new recommendation engine.
func NewEngine(config Config, store recmodel.EngineStore, rulesManager *rules.Manager, opts ...EngineOption) *Engine {
	engine := &Engine{
		config:       config,
		store:        store,
		rulesManager: rulesManager,
		clock:        realClock{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(engine)
		}
	}
	if engine.clock == nil {
		engine.clock = realClock{}
	}
	return engine
}

// Recommend generates recommendations using the blended scoring approach.
func (e *Engine) Recommend(
	ctx context.Context, req Request,
) (*Response, *TraceData, error) {
	req = e.sanitizeRequest(req)
	k := req.K

	sourceMetrics := make(map[string]SourceMetric)
	signalStatus := make(map[Signal]SignalStatus)

	usePrefetched := len(req.PrefetchedCandidates) > 0
	var popCandidates []recmodel.ScoredItem
	if usePrefetched {
		popCandidates = copyScoredItems(req.PrefetchedCandidates)
		sourceMetrics["prefetched"] = SourceMetric{Count: len(popCandidates)}
		e.recordSignalStatus(signalStatus, SignalPop, SignalOutcomeSuccess, false, nil)
	} else {
		popStart := time.Now()
		var err error
		popCandidates, err = e.getPopularityCandidates(
			ctx, req.OrgID, req.Namespace, k, req.Constraints,
		)
		if err != nil {
			e.recordSignalStatus(signalStatus, SignalPop, SignalOutcomeError, false, err)
			return nil, nil, err
		}
		sourceMetrics["popularity"] = SourceMetric{Count: len(popCandidates), Duration: time.Since(popStart)}
		e.recordSignalStatus(signalStatus, SignalPop, SignalOutcomeSuccess, false, nil)
	}

	existing := make(map[string]struct{}, len(popCandidates))
	maxPopScore := 0.0
	for _, c := range popCandidates {
		existing[c.ItemID] = struct{}{}
		if c.Score > maxPopScore {
			maxPopScore = c.Score
		}
	}

	if !usePrefetched && req.InjectAnchors && len(req.AnchorItemIDs) > 0 {
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
			popCandidates = append(popCandidates, recmodel.ScoredItem{ItemID: anchor, Score: maxPopScore})
			existing[anchor] = struct{}{}
		}
	}

	collabScores := make(map[string]float64)
	contentScores := make(map[string]float64)
	sessionScores := make(map[string]float64)
	if !usePrefetched && req.UserID != "" {
		start := time.Now()
		_, scores, err := e.getCollaborativeCandidates(ctx, req, existing, k)
		switch {
		case errors.Is(err, recmodel.ErrFeatureUnavailable):
			e.recordSignalStatus(signalStatus, SignalCollaborative, SignalOutcomeUnavailable, false, err)
			scores = nil
		case err != nil:
			e.recordSignalStatus(signalStatus, SignalCollaborative, SignalOutcomeError, false, err)
			return nil, nil, err
		default:
			e.recordSignalStatus(signalStatus, SignalCollaborative, SignalOutcomeSuccess, false, nil)
		}
		for id, score := range scores {
			collabScores[id] = score
		}
		sourceMetrics["collaborative"] = SourceMetric{Count: len(scores), Duration: time.Since(start)}

		start = time.Now()
		_, scores, err = e.getContentBasedCandidates(ctx, req, existing, k)
		switch {
		case errors.Is(err, recmodel.ErrFeatureUnavailable):
			e.recordSignalStatus(signalStatus, SignalContent, SignalOutcomeUnavailable, false, err)
			scores = nil
		case err != nil:
			e.recordSignalStatus(signalStatus, SignalContent, SignalOutcomeError, false, err)
			return nil, nil, err
		default:
			e.recordSignalStatus(signalStatus, SignalContent, SignalOutcomeSuccess, false, nil)
		}
		for id, score := range scores {
			contentScores[id] = score
		}
		sourceMetrics["content"] = SourceMetric{Count: len(scores), Duration: time.Since(start)}

		start = time.Now()
		_, scores, err = e.getSessionCandidates(ctx, req, existing, k)
		switch {
		case errors.Is(err, recmodel.ErrFeatureUnavailable):
			e.recordSignalStatus(signalStatus, SignalSession, SignalOutcomeUnavailable, false, err)
			scores = nil
		case err != nil:
			e.recordSignalStatus(signalStatus, SignalSession, SignalOutcomeError, false, err)
			return nil, nil, err
		default:
			e.recordSignalStatus(signalStatus, SignalSession, SignalOutcomeSuccess, false, nil)
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
	merged, sources := e.mergeCandidates(popCandidates, collabScores, contentScores, sessionScores, k)
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
	constraintStart := time.Now()
	candidates, tags = e.applyConstraintFilters(candidates, tags, req, &policySummary)
	sourceMetrics["post_constraints"] = SourceMetric{Count: len(candidates), Duration: time.Since(constraintStart)}

	// Align secondary score maps with filtered candidates.
	if len(popScores) > 0 {
		popScores = filterScoreMapByCandidates(popScores, candidates)
	}
	if len(collabScores) > 0 {
		collabScores = filterScoreMapByCandidates(collabScores, candidates)
	}
	if len(contentScores) > 0 {
		contentScores = filterScoreMapByCandidates(contentScores, candidates)
	}
	if len(sessionScores) > 0 {
		sessionScores = filterScoreMapByCandidates(sessionScores, candidates)
	}
	sources = filterSourcesByCandidates(sources, candidates)

	// Get blend weights.
	weights := e.getBlendWeights(req)

	// Gather co-visitation and embedding signals.
	candidateData, err := e.gatherSignals(ctx, candidates, req, weights, sources, signalStatus)
	if err != nil {
		return nil, nil, err
	}
	candidateData.Tags = tags
	candidateData.PopScores = popScores
	candidateData.CollabScores = collabScores
	candidateData.ContentScores = contentScores
	candidateData.SessionScores = sessionScores

	// Apply blended scoring.
	ApplyBlendedScoring(candidateData, weights)

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
	modelVersion := ModelVersionPopularity
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
		SignalStatus:       copySignalStatus(candidateData.SignalStatus),
		StarterProfile:     copyFloatMap(req.StarterProfile),
		StarterBlendWeight: req.StarterBlendWeight,
		RecentEventCount:   req.RecentEventCount,
	}
	var reasonSink map[string][]string
	if req.IncludeReasons {
		reasonSink = make(map[string][]string)
	}
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
	if ruleResult != nil && len(response.Items) > 0 {
		promoteManualBoosts(response, ruleResult)
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
		if len(ruleResult.OverrideHits) > 0 {
			hits := make([]rules.OverrideHit, len(ruleResult.OverrideHits))
			copy(hits, ruleResult.OverrideHits)
			trace.ManualOverrideHits = hits
		}
	}
	return response, trace, nil
}
