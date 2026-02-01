package mapper

import (
	"github.com/aatuh/recsys-suite/api/internal/services/recsysvc"
	"github.com/aatuh/recsys-suite/api/src/specs/types"
)

// NormalizedRecommendRequestDTO maps domain request to API normalized response.
func NormalizedRecommendRequestDTO(req recsysvc.RecommendRequest) types.NormalizedRecommendRequest {
	out := types.NormalizedRecommendRequest{
		Surface: req.Surface,
		Segment: req.Segment,
		K:       req.K,
		User: &types.UserRef{
			UserID:      req.User.UserID,
			AnonymousID: req.User.AnonymousID,
			SessionID:   req.User.SessionID,
		},
	}
	if req.Context != nil {
		out.Context = &types.RequestContext{
			Locale:  req.Context.Locale,
			Device:  req.Context.Device,
			Country: req.Context.Country,
			Now:     req.Context.Now,
		}
	}
	if req.Anchors != nil {
		out.Anchors = &types.AnchorsNormalized{
			ItemIDs:    append([]string(nil), req.Anchors.ItemIDs...),
			MaxAnchors: req.Anchors.MaxAnchors,
		}
	}
	if req.Candidates != nil {
		out.Candidates = &types.Candidates{
			IncludeIDs: append([]string(nil), req.Candidates.IncludeIDs...),
			ExcludeIDs: append([]string(nil), req.Candidates.ExcludeIDs...),
		}
	}
	if req.Constraints != nil {
		out.Constraints = &types.Constraints{
			RequiredTags:  append([]string(nil), req.Constraints.RequiredTags...),
			ForbiddenTags: append([]string(nil), req.Constraints.ForbiddenTags...),
			MaxPerTag:     cloneIntMap(req.Constraints.MaxPerTag),
		}
	}
	if req.Weights != nil {
		out.Weights = &types.Weights{Pop: req.Weights.Pop, Cooc: req.Weights.Cooc, Emb: req.Weights.Emb}
	}
	out.Options = &types.OptionsNormalized{
		IncludeReasons: req.Options.IncludeReasons,
		Explain:        req.Options.Explain,
		IncludeTrace:   req.Options.IncludeTrace,
		Seed:           req.Options.Seed,
	}
	if req.Experiment != nil {
		out.Experiment = &types.Experiment{ID: req.Experiment.ID, Variant: req.Experiment.Variant}
	}
	return out
}

// RecommendItemsDTO maps domain items to API items.
func RecommendItemsDTO(items []recsysvc.Item) []types.RecommendItem {
	if len(items) == 0 {
		return []types.RecommendItem{}
	}
	out := make([]types.RecommendItem, len(items))
	for i, it := range items {
		out[i] = types.RecommendItem{
			ItemID:  it.ItemID,
			Rank:    it.Rank,
			Score:   it.Score,
			Reasons: append([]string(nil), it.Reasons...),
		}
		if it.Explain != nil {
			out[i].Explain = &types.RecommendItemExplain{
				Signals: cloneFloatMap(it.Explain.Signals),
				Rules:   append([]string(nil), it.Explain.Rules...),
			}
		}
	}
	return out
}

// WarningsDTO maps domain warnings to API warnings.
func WarningsDTO(warnings []recsysvc.Warning) []types.Warning {
	if len(warnings) == 0 {
		return nil
	}
	out := make([]types.Warning, len(warnings))
	for i, w := range warnings {
		out[i] = types.Warning{Code: w.Code, Detail: w.Detail}
	}
	return out
}

func cloneIntMap(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneFloatMap(in map[string]float64) map[string]float64 {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]float64, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
