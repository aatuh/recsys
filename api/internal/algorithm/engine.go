package algorithm

import (
	"context"
	"fmt"
	"math"
	"time"

	"recsys/internal/types"

	"github.com/google/uuid"
)

// Default recent anchors window days if not configured.
const DEFAULT_RECENT_ANCHORS_WINDOW_DAYS = 30

// Default co-visitation window days if not configured.
const DEFAULT_COVIS_WINDOW_DAYS = 30

// Limit anchors for performance.
const MAX_RECENT_ANCHORS = 10

// Limit co-visitation neighbors for performance.
const MAX_COVIS_NEIGHBORS = 200

// Limit embedding neighbors for performance.
const MAX_EMB_NEIGHBORS = 200

// Engine handles recommendation algorithm logic
type Engine struct {
	config Config
	store  types.AlgoStore
}

// NewEngine creates a new recommendation engine
func NewEngine(config Config, store types.AlgoStore) *Engine {
	fmt.Println("config", config)
	return &Engine{
		config: config,
		store:  store,
	}
}

// Recommend generates recommendations using the blended scoring approach
func (e *Engine) Recommend(
	ctx context.Context, req Request,
) (*Response, error) {
	// Set defaults
	k := req.K
	if k <= 0 {
		k = 20
	}

	// Get popularity candidates
	candidates, err := e.getPopularityCandidates(
		ctx, req.OrgID, req.Namespace, k, req.Constraints,
	)
	if err != nil {
		return nil, err
	}

	// Apply exclusions
	candidates, err = e.applyExclusions(ctx, candidates, req)
	if err != nil {
		return nil, err
	}

	// Get tags for all candidates
	tags, err := e.getCandidateTags(ctx, candidates, req)
	if err != nil {
		return nil, err
	}

	// Get blend weights
	weights := e.getBlendWeights(req)

	// Gather co-visitation and embedding signals
	candidateData, err := e.gatherSignals(ctx, candidates, req, weights)
	if err != nil {
		return nil, err
	}
	candidateData.Tags = tags

	// Apply blended scoring
	e.applyBlendedScoring(candidateData, weights)

	// Apply personalization boost
	e.applyPersonalizationBoost(ctx, candidateData, req)

	// Determine model version
	modelVersion := e.getModelVersion(weights)

	// Apply MMR and caps if needed
	if e.shouldUseMMR() || e.shouldUseCaps() {
		candidateData.Candidates = e.applyMMRAndCaps(candidateData, k)
	}

	// Build response
	return e.buildResponse(
		candidateData, k, modelVersion, req.IncludeReasons, weights,
	), nil
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

	// Fetch popularity candidates
	cands, err := e.store.PopularityTopK(
		ctx, orgID, ns, e.config.HalfLifeDays, fetchK, c,
	)
	if err != nil {
		return nil, err
	}

	return cands, nil
}

// applyExclusions removes excluded items from candidates
func (e *Engine) applyExclusions(
	ctx context.Context,
	candidates []types.ScoredItem,
	req Request,
) ([]types.ScoredItem, error) {
	exclude := make(map[string]struct{})

	// Add constraint exclusions
	if req.Constraints != nil {
		for _, id := range req.Constraints.ExcludeItemIDs {
			exclude[id] = struct{}{}
		}
	}

	// Add purchased exclusions if enabled
	var err error
	exclude, err = e.excludePurchasedItems(ctx, req, exclude)
	if err != nil {
		return nil, err
	}

	// Filter candidates by excluding excluded items
	filtered := make([]types.ScoredItem, 0, len(candidates))
	for _, candidate := range candidates {
		if _, skip := exclude[candidate.ItemID]; !skip {
			filtered = append(filtered, candidate)
		}
	}

	return filtered, nil
}

// excludePurchasedItems excludes purchased items from the candidate set.
func (e *Engine) excludePurchasedItems(
	ctx context.Context,
	req Request,
	exclude map[string]struct{},
) (map[string]struct{}, error) {
	if !e.config.RuleExcludePurchased || req.UserID == "" {
		return exclude, nil
	}

	// Exclude purchased items in a time window
	lookback := time.Duration(e.config.PurchasedWindowDays*24.0) * time.Hour
	since := time.Now().UTC().Add(-lookback)
	bought, err := e.store.ListUserPurchasedSince(
		ctx,
		req.OrgID,
		req.Namespace,
		req.UserID,
		since,
	)
	if err != nil {
		return nil, err
	}

	// Add purchased items to exclude
	for _, id := range bought {
		exclude[id] = struct{}{}
	}

	return exclude, nil
}

// getCandidateTags fetches tags for all candidates
func (e *Engine) getCandidateTags(
	ctx context.Context, candidates []types.ScoredItem, req Request,
) (map[string]types.ItemTags, error) {
	if len(candidates) == 0 {
		return make(map[string]types.ItemTags), nil
	}

	// Build list of candidate IDs
	ids := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		ids = append(ids, candidate.ItemID)
	}

	// Fetch tags for all candidates
	return e.store.ListItemsTags(ctx, req.OrgID, req.Namespace, ids)
}

// getBlendWeights returns the blend weights to use
func (e *Engine) getBlendWeights(req Request) BlendWeights {
	weights := BlendWeights{
		Pop:  e.config.BlendAlpha,
		Cooc: e.config.BlendBeta,
		ALS:  e.config.BlendGamma,
	}

	// Override with request weights if provided
	if req.Blend != nil {
		weights.Pop = math.Max(0, req.Blend.Pop)
		weights.Cooc = math.Max(0, req.Blend.Cooc)
		weights.ALS = math.Max(0, req.Blend.ALS)
	}

	// Safety: if all weights are 0, use popularity-only
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
		Candidates: candidates,
		CoocScores: make(map[string]float64),
		EmbScores:  make(map[string]float64),
		UsedCooc:   make(map[string]bool),
		UsedEmb:    make(map[string]bool),
		Boosted:    make(map[string]bool),
	}

	// Build candidate set for quick lookup
	candSet := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		candSet[candidate.ItemID] = struct{}{}
	}

	// Only gather signals if user ID is provided and weights are non-zero
	if req.UserID == "" || (weights.Cooc == 0 && weights.ALS == 0) {
		return data, nil
	}

	// Get recent anchors
	anchors, err := e.getRecentAnchors(ctx, req)
	if err != nil || len(anchors) == 0 {
		return data, nil
	}

	// Gather co-visitation signals
	if weights.Cooc > 0 {
		err = e.gatherCoVisitationSignals(ctx, data, req, anchors, candSet)
		if err != nil {
			return nil, err
		}
	}

	// Gather embedding signals
	if weights.ALS > 0 {
		err = e.gatherEmbeddingSignals(ctx, data, req, anchors, candSet)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// getRecentAnchors gets recent items for the user to use as anchors
func (e *Engine) getRecentAnchors(
	ctx context.Context, req Request,
) ([]string, error) {
	days := e.config.CoVisWindowDays
	if days <= 0 {
		days = DEFAULT_RECENT_ANCHORS_WINDOW_DAYS
	}
	since := time.Now().UTC().Add(-time.Duration(days*24.0) * time.Hour)

	return e.store.ListUserRecentItemIDs(
		ctx,
		req.OrgID,
		req.Namespace,
		req.UserID,
		since,
		MAX_RECENT_ANCHORS,
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
		days = DEFAULT_COVIS_WINDOW_DAYS
	}
	since := time.Now().UTC().Add(-time.Duration(days*24.0) * time.Hour)

	for _, anchor := range anchors {
		// Gather co-visitation neighbors
		neighbors, err := e.store.CooccurrenceTopKWithin(
			ctx,
			req.OrgID,
			req.Namespace,
			anchor,
			MAX_COVIS_NEIGHBORS,
			since,
		)
		if err != nil {
			continue // Skip this anchor on error
		}

		// Update co-visitation scores.
		for _, neighbor := range neighbors {
			if _, ok := candSet[neighbor.ItemID]; !ok {
				continue // Not in candidate set
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
		// Gather embedding neighbors
		neighbors, err := e.store.SimilarByEmbeddingTopK(
			ctx,
			req.OrgID,
			req.Namespace,
			anchor,
			MAX_EMB_NEIGHBORS,
		)
		if err != nil {
			continue // Skip this anchor on error
		}

		// Update embedding scores.
		for _, neighbor := range neighbors {
			if _, ok := candSet[neighbor.ItemID]; !ok {
				continue // Not in candidate set
			}
			// Set score if higher than current score.
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
	// Find max scores for normalization
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

	// Apply blended scoring
	for i := range data.Candidates {
		id := data.Candidates[i].ItemID

		popNorm := 0.0
		if maxPop > 0 {
			popNorm = data.Candidates[i].Score / maxPop
		}

		coocNorm := 0.0
		if maxCooc > 0 {
			coocNorm = data.CoocScores[id] / maxCooc
		}

		embNorm := 0.0
		if maxEmb > 0 {
			embNorm = data.EmbScores[id] / maxEmb
		}

		blended := weights.Pop*popNorm +
			weights.Cooc*coocNorm + weights.ALS*embNorm

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

	profile, err := e.store.BuildUserTagProfile(
		ctx,
		req.OrgID,
		req.Namespace,
		req.UserID,
		e.config.ProfileWindowDays,
		maxInt(e.config.ProfileTopNTags, 1),
	)
	if err != nil || len(profile) == 0 {
		return
	}

	for i := range data.Candidates {
		id := data.Candidates[i].ItemID
		tags := data.Tags[id]

		overlap := 0.0
		for _, tag := range tags.Tags {
			if weight, ok := profile[tag]; ok {
				overlap += weight
			}
		}

		if overlap > 0 {
			data.Candidates[i].Score *= (1.0 + e.config.ProfileBoost*overlap)
			data.Boosted[id] = true
		}
	}
}

// getModelVersion determines the model version based on weights.
func (e *Engine) getModelVersion(weights BlendWeights) string {
	if weights.Cooc == 0 && weights.ALS == 0 {
		return "popularity_v1"
	}
	return "blend_v1"
}

// shouldUseMMR returns true if MMR should be applied.
func (e *Engine) shouldUseMMR() bool {
	return e.config.MMRLambda > 0
}

// shouldUseCaps returns true if caps should be applied.
func (e *Engine) shouldUseCaps() bool {
	return e.config.BrandCap > 0 || e.config.CategoryCap > 0
}

// applyMMRAndCaps applies MMR re-ranking and caps.
func (e *Engine) applyMMRAndCaps(
	data *CandidateData, k int,
) []types.ScoredItem {
	return MMRReRank(
		data.Candidates,
		data.Tags,
		k,
		e.config.MMRLambda,
		e.config.BrandCap,
		e.config.CategoryCap,
	)
}

// buildResponse builds the final response.
func (e *Engine) buildResponse(
	data *CandidateData,
	k int,
	modelVersion string,
	includeReasons bool,
	weights BlendWeights,
) *Response {
	response := &Response{
		ModelVersion: modelVersion,
		Items:        make([]ScoredItem, 0, minInt(k, len(data.Candidates))),
	}

	for i, candidate := range data.Candidates {
		if i >= k {
			break
		}

		reasons := e.buildReasons(
			candidate.ItemID, includeReasons, weights, data,
		)

		response.Items = append(response.Items, ScoredItem{
			ItemID:  candidate.ItemID,
			Score:   candidate.Score,
			Reasons: reasons,
		})
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

// Utility functions.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

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
