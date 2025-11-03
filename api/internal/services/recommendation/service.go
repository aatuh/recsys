package recommendation

import (
	"context"
	"errors"
	"strings"
	"time"

	"recsys/internal/algorithm"
	"recsys/internal/rules"
	"recsys/internal/types"
	spectypes "recsys/specs/types"

	"github.com/google/uuid"
)

// ValidationError represents client-side validation failures.
type ValidationError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// Store defines the persistence contract needed by the service.
type Store interface {
	types.RecAlgoStore
}

// SegmentSelection describes the resolved segment context.
type SegmentSelection struct {
	Profile     *types.SegmentProfile
	SegmentID   string
	ProfileID   string
	RuleID      int64
	UserTraits  map[string]any
	UserCreated time.Time
}

// SegmentSelector resolves segment information prior to ranking.
type SegmentSelector func(ctx context.Context, req algorithm.Request, httpReq spectypes.RecommendRequest) (SegmentSelection, error)

// Result encapsulates recommendation outputs for the HTTP layer.
type Result struct {
	Response     spectypes.RecommendResponse
	AlgoRequest  algorithm.Request
	AlgoConfig   algorithm.Config
	AlgoResponse *algorithm.Response
	TraceData    *algorithm.TraceData
	SourceStats  map[string]algorithm.SourceMetric
}

// Service orchestrates recommendation requests.
type Service struct {
	store         Store
	rules         *rules.Manager
	blendResolver BlendConfigResolver
}

// New constructs a recommendation service.
func New(store Store, rulesManager *rules.Manager) *Service {
	return &Service{store: store, rules: rulesManager}
}

// WithBlendResolver configures a resolver for runtime blend overrides.
func (s *Service) WithBlendResolver(resolver BlendConfigResolver) *Service {
	s.blendResolver = resolver
	return s
}

// Recommend executes the ranking pipeline using the provided context.
func (s *Service) Recommend(
	ctx context.Context,
	orgID uuid.UUID,
	req spectypes.RecommendRequest,
	baseCfg algorithm.Config,
	selector SegmentSelector,
) (*Result, error) {
	algoReq, err := buildAlgorithmRequest(orgID, req)
	if err != nil {
		return nil, err
	}
	cfg := baseCfg

	var selection SegmentSelection
	if selector != nil {
		sel, err := selector(ctx, algoReq, req)
		if err != nil {
			return nil, err
		}
		selection = sel
		if selection.Profile != nil {
			applySegmentProfile(&cfg, *selection.Profile)
			if selection.ProfileID == "" {
				selection.ProfileID = selection.Profile.ProfileID
			}
		}
		algoReq.SegmentID = selection.SegmentID
	}

	blendOverrides(&cfg, s.resolveBlend(ctx, algoReq.Namespace))
	applyOverrides(&cfg, req.Overrides)

	isNewSegment := strings.EqualFold(selection.SegmentID, "new_users")
	if !isNewSegment && selection.UserTraits != nil {
		if seg, ok := selection.UserTraits["segment"].(string); ok {
			isNewSegment = strings.EqualFold(seg, "new_users")
		}
	}
	if isNewSegment {
		cfg.RuleExcludeEvents = false
		if cfg.PopularityFanout < 1000 {
			cfg.PopularityFanout = 1000
		}
	}

	recentItems, err := s.recentInteractionItems(ctx, cfg, algoReq)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		recentItems = nil
	}

	recentEventCount := len(recentItems)
	recentCountKnown := recentItems != nil
	effectiveEventCount := recentEventCount
	if !recentCountKnown {
		effectiveEventCount = -1
	}
	if isNewSegment && cfg.ProfileMinEventsForBoost > 0 && effectiveEventCount >= cfg.ProfileMinEventsForBoost {
		effectiveEventCount = cfg.ProfileMinEventsForBoost - 1
		if effectiveEventCount < 0 {
			effectiveEventCount = 0
		}
	}
	algoReq.RecentEventCount = effectiveEventCount

	if starter := s.buildStarterProfile(ctx, cfg, algoReq, selection, recentEventCount, recentCountKnown, recentItems); len(starter) > 0 {
		algoReq.StarterProfile = starter
		algoReq.StarterBlendWeight = cfg.ProfileStarterBlendWeight
	}

	engine := algorithm.NewEngine(cfg, s.store, s.rules)
	algoResp, traceData, err := engine.Recommend(ctx, algoReq)
	if err != nil {
		return nil, err
	}

	if selector != nil {
		selection, _ := selector(ctx, algoReq, req)
		if selection.SegmentID != "" {
			algoResp.SegmentID = selection.SegmentID
		}
		if selection.ProfileID != "" {
			algoResp.ProfileID = selection.ProfileID
		}
	}

	httpResp := convertToHTTPResponse(algoResp)

	return &Result{
		Response:     httpResp,
		AlgoRequest:  algoReq,
		AlgoConfig:   cfg,
		AlgoResponse: algoResp,
		TraceData:    traceData,
		SourceStats:  traceData.SourceMetrics,
	}, nil
}

func buildAlgorithmRequest(orgID uuid.UUID, req spectypes.RecommendRequest) (algorithm.Request, error) {
	ns := req.Namespace
	if ns == "" {
		ns = "default"
	}

	constraints, err := parseConstraints(req.Constraints)
	if err != nil {
		return algorithm.Request{}, err
	}

	return algorithm.Request{
		OrgID:          orgID,
		UserID:         req.UserID,
		Namespace:      ns,
		K:              req.K,
		Blend:          parseBlend(req.Blend),
		Constraints:    constraints,
		IncludeReasons: req.IncludeReasons,
		ExplainLevel:   algorithm.NormalizeExplainLevel(req.ExplainLevel),
		Surface:        extractSurface(req.Context),
	}, nil
}

func extractSurface(ctx map[string]any) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx["surface"]; ok {
		if s, ok := val.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func parseConstraints(src *spectypes.RecommendConstraints) (*types.PopConstraints, error) {
	if src == nil {
		return nil, nil
	}
	constraints := &types.PopConstraints{
		IncludeTagsAny: src.IncludeTagsAny,
		ExcludeItemIDs: src.ExcludeItemIDs,
	}
	if len(src.PriceBetween) >= 1 {
		v := src.PriceBetween[0]
		constraints.MinPrice = &v
	}
	if len(src.PriceBetween) >= 2 {
		v := src.PriceBetween[1]
		constraints.MaxPrice = &v
	}
	if src.CreatedAfterISO != "" {
		ts, err := time.Parse(time.RFC3339, src.CreatedAfterISO)
		if err != nil {
			return nil, ValidationError{Code: "invalid_created_after", Message: "created_after must be RFC3339"}
		}
		constraints.CreatedAfter = &ts
	}
	return constraints, nil
}

func parseBlend(src *spectypes.RecommendBlend) *algorithm.BlendWeights {
	if src == nil {
		return nil
	}
	return &algorithm.BlendWeights{
		Pop:  src.Pop,
		Cooc: src.Cooc,
		ALS:  src.ALS,
	}
}

func applyOverrides(cfg *algorithm.Config, overrides *spectypes.Overrides) {
	if overrides == nil {
		return
	}
	if overrides.BlendAlpha != nil {
		cfg.BlendAlpha = *overrides.BlendAlpha
	}
	if overrides.BlendBeta != nil {
		cfg.BlendBeta = *overrides.BlendBeta
	}
	if overrides.BlendGamma != nil {
		cfg.BlendGamma = *overrides.BlendGamma
	}
	if overrides.ProfileBoost != nil {
		cfg.ProfileBoost = *overrides.ProfileBoost
	}
	if overrides.ProfileWindowDays != nil {
		cfg.ProfileWindowDays = float64(*overrides.ProfileWindowDays)
	}
	if overrides.ProfileTopN != nil {
		cfg.ProfileTopNTags = *overrides.ProfileTopN
	}
	if overrides.ProfileStarterBlendWeight != nil {
		weight := *overrides.ProfileStarterBlendWeight
		if weight < 0 {
			weight = 0
		} else if weight > 1 {
			weight = 1
		}
		cfg.ProfileStarterBlendWeight = weight
	}
	if overrides.MMRLambda != nil {
		cfg.MMRLambda = *overrides.MMRLambda
	}
	if overrides.BrandCap != nil {
		cfg.BrandCap = *overrides.BrandCap
	}
	if overrides.CategoryCap != nil {
		cfg.CategoryCap = *overrides.CategoryCap
	}
	if overrides.PopularityHalfLifeDays != nil {
		cfg.HalfLifeDays = float64(*overrides.PopularityHalfLifeDays)
	}
	if overrides.CoVisWindowDays != nil {
		cfg.CoVisWindowDays = *overrides.CoVisWindowDays
	}
	if overrides.PurchasedWindowDays != nil {
		cfg.PurchasedWindowDays = *overrides.PurchasedWindowDays
	}
	if overrides.RuleExcludeEvents != nil {
		cfg.RuleExcludeEvents = *overrides.RuleExcludeEvents
	}
	if overrides.PopularityFanout != nil {
		cfg.PopularityFanout = *overrides.PopularityFanout
	}
}

func (s *Service) resolveBlend(ctx context.Context, namespace string) *ResolvedBlendConfig {
	if s.blendResolver == nil {
		return nil
	}
	resolved, err := s.blendResolver.ResolveBlend(ctx, normalizeNamespace(namespace))
	if err != nil {
		return nil
	}
	return resolved
}

func applySegmentProfile(cfg *algorithm.Config, profile types.SegmentProfile) {
	cfg.BlendAlpha = profile.BlendAlpha
	cfg.BlendBeta = profile.BlendBeta
	cfg.BlendGamma = profile.BlendGamma
	cfg.MMRLambda = profile.MMRLambda
	cfg.BrandCap = profile.BrandCap
	cfg.CategoryCap = profile.CategoryCap
	cfg.ProfileBoost = profile.ProfileBoost
	cfg.ProfileWindowDays = profile.ProfileWindowDays
	cfg.ProfileTopNTags = profile.ProfileTopN
	cfg.HalfLifeDays = profile.HalfLifeDays
	cfg.CoVisWindowDays = profile.CoVisWindowDays
	cfg.PurchasedWindowDays = profile.PurchasedWindowDays
	cfg.RuleExcludeEvents = profile.RuleExcludeEvents
	cfg.ExcludeEventTypes = append([]int16(nil), profile.ExcludeEventTypes...)
	cfg.BrandTagPrefixes = append([]string(nil), profile.BrandTagPrefixes...)
	cfg.CategoryTagPrefixes = append([]string(nil), profile.CategoryTagPrefixes...)
	cfg.PopularityFanout = profile.PopularityFanout
}

func convertToHTTPResponse(algoResp *algorithm.Response) spectypes.RecommendResponse {
	items := make([]spectypes.ScoredItem, 0, len(algoResp.Items))
	for _, item := range algoResp.Items {
		var explain *spectypes.ExplainBlock
		if item.Explain != nil {
			explain = mapExplainBlock(item.Explain)
		}
		items = append(items, spectypes.ScoredItem{
			ItemID:  item.ItemID,
			Score:   item.Score,
			Reasons: item.Reasons,
			Explain: explain,
		})
	}

	return spectypes.RecommendResponse{
		ModelVersion: algoResp.ModelVersion,
		Items:        items,
		SegmentID:    algoResp.SegmentID,
		ProfileID:    algoResp.ProfileID,
	}
}

// BuildAlgorithmRequest exposes request conversion for other handlers.
func BuildAlgorithmRequest(orgID uuid.UUID, req spectypes.RecommendRequest) (algorithm.Request, error) {
	return buildAlgorithmRequest(orgID, req)
}

// ApplyOverrides exposes override application for other handlers.
func ApplyOverrides(cfg *algorithm.Config, overrides *spectypes.Overrides) {
	applyOverrides(cfg, overrides)
}

// ConvertToHTTPResponse exposes response conversion for other handlers.
func ConvertToHTTPResponse(resp *algorithm.Response) spectypes.RecommendResponse {
	return convertToHTTPResponse(resp)
}

func mapExplainBlock(src *algorithm.ExplainBlock) *spectypes.ExplainBlock {
	if src == nil {
		return nil
	}
	dst := &spectypes.ExplainBlock{}

	if src.Blend != nil {
		dst.Blend = &spectypes.ExplainBlend{
			Alpha:    src.Blend.Alpha,
			Beta:     src.Blend.Beta,
			Gamma:    src.Blend.Gamma,
			PopNorm:  src.Blend.PopNorm,
			CoocNorm: src.Blend.CoocNorm,
			EmbNorm:  src.Blend.EmbNorm,
			Contributions: spectypes.ExplainBlendContribution{
				Pop:  src.Blend.Contributions.Pop,
				Cooc: src.Blend.Contributions.Cooc,
				Emb:  src.Blend.Contributions.Emb,
			},
		}
		if src.Blend.Raw != nil {
			dst.Blend.Raw = &spectypes.ExplainBlendRaw{
				Pop:  src.Blend.Raw.Pop,
				Cooc: src.Blend.Raw.Cooc,
				Emb:  src.Blend.Raw.Emb,
			}
		}
	}

	if src.Personalization != nil {
		dst.Personalization = &spectypes.ExplainPersonalization{
			Overlap:         src.Personalization.Overlap,
			BoostMultiplier: src.Personalization.BoostMultiplier,
		}
	}

	if src.MMR != nil {
		dst.MMR = &spectypes.ExplainMMR{
			Lambda:        src.MMR.Lambda,
			MaxSimilarity: src.MMR.MaxSimilarity,
			Penalty:       src.MMR.Penalty,
			Relevance:     src.MMR.Relevance,
			Rank:          src.MMR.Rank,
		}
	}

	if src.Caps != nil {
		caps := &spectypes.ExplainCaps{}
		if src.Caps.Brand != nil {
			caps.Brand = mapCapUsage(src.Caps.Brand)
		}
		if src.Caps.Category != nil {
			caps.Category = mapCapUsage(src.Caps.Category)
		}
		if caps.Brand != nil || caps.Category != nil {
			dst.Caps = caps
		}
	}

	if len(src.Anchors) > 0 {
		dst.Anchors = append([]string(nil), src.Anchors...)
	}

	if dst.Blend == nil && dst.Personalization == nil && dst.MMR == nil && dst.Caps == nil && len(dst.Anchors) == 0 {
		return nil
	}
	return dst
}

func mapCapUsage(src *algorithm.CapUsage) *spectypes.ExplainCapUsage {
	if src == nil {
		return nil
	}
	usage := &spectypes.ExplainCapUsage{
		Applied: src.Applied,
		Value:   src.Value,
	}
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
