package algorithm

import (
	"sort"

	"github.com/aatuh/recsys-algo/rules"

	recmodel "github.com/aatuh/recsys-algo/model"
)

// getModelVersion determines the model version based on weights.
func (e *Engine) getModelVersion(weights BlendWeights) string {
	return ModelVersionForWeights(weights)
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
	needReasons := includeReasons || reasonSink != nil

	// Attach pinned items first.
	for _, pin := range pinned {
		if remaining <= 0 {
			break
		}
		var reasonsFull []string
		if needReasons {
			reasonsFull = appendReasons(pin.ItemID, e.buildReasons(pin.ItemID, true, weights, data))
			if reasonSink != nil {
				reasonSink[pin.ItemID] = append([]string(nil), reasonsFull...)
			}
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
	sortedCandidates := make([]recmodel.ScoredItem, len(data.Candidates))
	copy(sortedCandidates, data.Candidates)
	sort.SliceStable(sortedCandidates, func(i, j int) bool {
		if sortedCandidates[i].Score == sortedCandidates[j].Score {
			return sortedCandidates[i].ItemID < sortedCandidates[j].ItemID
		}
		return sortedCandidates[i].Score > sortedCandidates[j].Score
	})

	for _, candidate := range sortedCandidates {
		if remaining <= 0 {
			break
		}
		var reasonsFull []string
		if needReasons {
			reasonsFull = appendReasons(candidate.ItemID, e.buildReasons(candidate.ItemID, true, weights, data))
			if reasonSink != nil {
				reasonSink[candidate.ItemID] = append([]string(nil), reasonsFull...)
			}
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

	if weights.Pop > 0 && data.PopRaw[itemID] > 0 {
		reasons = append(reasons, "recent_popularity")
	}
	if weights.Cooc > 0 && data.CoocRaw[itemID] > 0 {
		reasons = append(reasons, "co_visitation")
	}

	if weights.Similarity > 0 {
		if sources := data.SimilaritySources[itemID]; len(sources) > 0 {
			for _, source := range sources {
				switch source {
				case SignalPop:
					reasons = append(reasons, "recent_popularity")
				case SignalCooc:
					reasons = append(reasons, "co_visitation")
				case SignalEmbedding:
					reasons = append(reasons, "embedding_similarity")
				case SignalCollaborative:
					reasons = append(reasons, "collaborative")
				case SignalContent:
					reasons = append(reasons, "content_similarity")
				case SignalSession:
					reasons = append(reasons, "session_sequence")
				}
			}
		}
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

	if weights.Pop > 0 || weights.Cooc > 0 || weights.Similarity > 0 {
		blend := &BlendExplain{
			Alpha:          weights.Pop,
			Beta:           weights.Cooc,
			Gamma:          weights.Similarity,
			PopNorm:        data.PopNorm[itemID],
			CoocNorm:       data.CoocNorm[itemID],
			SimilarityNorm: data.SimilarityNorm[itemID],
			Contributions: BlendContribution{
				Pop:        weights.Pop * data.PopNorm[itemID],
				Cooc:       weights.Cooc * data.CoocNorm[itemID],
				Similarity: weights.Similarity * data.SimilarityNorm[itemID],
			},
		}
		if level == ExplainLevelFull {
			blend.Raw = &BlendRaw{
				Pop:        data.PopRaw[itemID],
				Cooc:       data.CoocRaw[itemID],
				Similarity: data.SimilarityRaw[itemID],
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
				if summary.ConstraintLeakByReason == nil {
					summary.ConstraintLeakByReason = make(map[string]int)
				} else {
					for k := range summary.ConstraintLeakByReason {
						delete(summary.ConstraintLeakByReason, k)
					}
				}
				if len(leaks) > maxPolicySampleIDs {
					summary.ConstraintLeakIDs = append([]string(nil), leaks[:maxPolicySampleIDs]...)
				} else {
					summary.ConstraintLeakIDs = append([]string(nil), leaks...)
				}
				for _, leaked := range leaks {
					reason := constraintReasonUnknown
					if summary.constraintFilteredReasons != nil {
						if r, ok := summary.constraintFilteredReasons[leaked]; ok && r != "" {
							reason = r
						}
					}
					summary.ConstraintLeakByReason[reason]++
				}
			} else {
				summary.ConstraintLeakIDs = nil
				summary.ConstraintLeakByReason = nil
			}
		} else {
			summary.ConstraintLeakCount = 0
			summary.ConstraintLeakIDs = nil
			summary.ConstraintLeakByReason = nil
		}
		summary.constraintFilteredLookup = nil
		summary.constraintFilteredReasons = nil

		if ruleResult != nil && len(ruleResult.ItemEffects) > 0 {
			boostExposure := 0
			pinExposure := 0
			for _, item := range resp.Items {
				if eff, ok := ruleResult.ItemEffects[item.ItemID]; ok {
					if eff.BoostDelta != 0 {
						boostExposure++
						for _, boost := range eff.BoostRules {
							if hit := ruleResult.OverrideHitForRule(boost.RuleID); hit != nil {
								hit.ServedItems = appendUniqueString(hit.ServedItems, item.ItemID)
							}
						}
					}
					if eff.Pinned {
						pinExposure++
						for _, ruleID := range eff.PinRules {
							if hit := ruleResult.OverrideHitForRule(ruleID); hit != nil {
								hit.ServedItems = appendUniqueString(hit.ServedItems, item.ItemID)
							}
						}
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
	if ruleResult != nil && len(ruleResult.ItemEffects) > 0 {
		blockExposure := 0
		blockByRule := make(map[string]int)
		for _, eff := range ruleResult.ItemEffects {
			if !eff.Blocked {
				continue
			}
			blockExposure++
			if len(eff.BlockRules) == 0 {
				blockByRule[constraintReasonUnknown]++
				continue
			}
			for _, ruleID := range eff.BlockRules {
				blockByRule[ruleID.String()]++
			}
		}
		summary.RuleBlockExposure = blockExposure
		if blockExposure > 0 {
			summary.RuleBlockExposureByRule = blockByRule
		} else {
			summary.RuleBlockExposureByRule = nil
		}

		if summary.RuleBoostExposure == 0 && summary.RulePinExposure == 0 {
			for _, hit := range ruleResult.OverrideHits {
				if len(hit.ServedItems) > 0 {
					summary.RuleBoostExposure = len(hit.ServedItems)
				}
			}
		}
	} else {
		summary.RuleBlockExposure = 0
		summary.RuleBlockExposureByRule = nil
	}
}

func appendUniqueString(dst []string, value string) []string {
	for _, existing := range dst {
		if existing == value {
			return dst
		}
	}
	return append(dst, value)
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
