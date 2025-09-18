package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"time"

	"recsys/internal/algorithm"
	"recsys/internal/audit"
	handlerstypes "recsys/internal/http/types"
	"recsys/internal/types"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type decisionTraceInput struct {
	Request      *http.Request
	HTTPRequest  handlerstypes.RecommendRequest
	AlgoRequest  algorithm.Request
	Config       algorithm.Config
	AlgoResponse *algorithm.Response
	HTTPResponse handlerstypes.RecommendResponse
	TraceData    *algorithm.TraceData
	Duration     time.Duration
	Surface      string
	Bandit       *banditTraceContext
}

type banditTraceContext struct {
	ChosenPolicyID string
	Algorithm      string
	BucketKey      string
	Explore        bool
	RequestID      string
	Explain        map[string]string
}

func (h *Handler) recordDecisionTrace(in decisionTraceInput) {
	if h.DecisionRecorder == nil {
		if h.Logger != nil {
			h.auditRecorderWarnOnce.Do(func() {
				h.Logger.Warn("decision recorder not configured; skipping audit trace")
			})
		}
		return
	}
	if in.TraceData == nil {
		return
	}

	namespace := in.AlgoRequest.Namespace
	if namespace == "" {
		namespace = "default"
	}

	trace := audit.Trace{
		DecisionID: uuid.New(),
		OrgID:      in.AlgoRequest.OrgID.String(),
		Timestamp:  time.Now().UTC(),
		Namespace:  namespace,
		Surface:    in.Surface,
		K:          in.TraceData.K,
		Config: audit.TraceConfig{
			Alpha:               in.Config.BlendAlpha,
			Beta:                in.Config.BlendBeta,
			Gamma:               in.Config.BlendGamma,
			ProfileBoost:        in.Config.ProfileBoost,
			ProfileWindowDays:   in.Config.ProfileWindowDays,
			ProfileTopN:         in.Config.ProfileTopNTags,
			MMRLambda:           in.Config.MMRLambda,
			BrandCap:            in.Config.BrandCap,
			CategoryCap:         in.Config.CategoryCap,
			HalfLifeDays:        in.Config.HalfLifeDays,
			CoVisWindowDays:     in.Config.CoVisWindowDays,
			PurchasedWindowDays: in.Config.PurchasedWindowDays,
			RuleExcludeEvents:   in.Config.RuleExcludeEvents,
			PopularityFanout:    in.Config.PopularityFanout,
		},
	}

	if reqID := middleware.GetReqID(in.Request.Context()); reqID != "" {
		trace.RequestID = reqID
	}
	if hash := hashUser(namespace, in.AlgoRequest.UserID, h.DecisionTraceSalt); hash != "" {
		trace.UserHash = hash
	}

	if c := buildTraceConstraints(in.AlgoRequest.Constraints); c != nil {
		trace.Constraints = c
	}

	trace.Candidates = buildTraceCandidates(in.TraceData.CandidatesPre)
	trace.FinalItems = buildTraceFinalItems(in.HTTPResponse.Items, in.TraceData.Reasons)
	trace.MMR = buildTraceMMR(in.TraceData.MMRInfo, in.TraceData.CapsInfo)
	if caps := buildTraceCaps(in.TraceData.CapsInfo); len(caps) > 0 {
		trace.Caps = caps
	}

	extras := map[string]any{
		"model_version":   in.TraceData.ModelVersion,
		"duration_ms":     float64(in.Duration.Milliseconds()),
		"include_reasons": in.TraceData.IncludeReasons,
		"explain_level":   string(in.TraceData.ExplainLevel),
	}

	if len(in.TraceData.Anchors) > 0 {
		anchors := append([]string(nil), in.TraceData.Anchors...)
		extras["anchors"] = anchors
	}
	if seg := in.AlgoResponse.SegmentID; seg != "" {
		extras["segment_id"] = seg
	}
	if profile := in.AlgoResponse.ProfileID; profile != "" {
		extras["profile_id"] = profile
	}
	if boosted := buildBoostedList(in.TraceData.Boosted); len(boosted) > 0 {
		extras["personalized_items"] = boosted
	}
	if ctxKeys := extractContextKeys(in.HTTPRequest.Context); len(ctxKeys) > 0 {
		extras["request_context_keys"] = ctxKeys
	}
	if len(in.TraceData.MMRInfo) > 0 {
		extras["mmr_applied"] = true
	}
	if len(in.TraceData.CapsInfo) > 0 {
		extras["caps_applied"] = true
	}

	if len(extras) > 0 {
		trace.Extras = extras
	}

	if in.Bandit != nil {
		trace.Bandit = &audit.TraceBandit{
			ChosenPolicyID: in.Bandit.ChosenPolicyID,
			Algorithm:      in.Bandit.Algorithm,
			BucketKey:      in.Bandit.BucketKey,
			Explore:        in.Bandit.Explore,
			RequestID:      in.Bandit.RequestID,
			Explain:        in.Bandit.Explain,
		}
	}

	if h.Logger != nil {
		h.Logger.Debug("queueing decision trace", zap.String("decision_id", trace.DecisionID.String()), zap.String("namespace", trace.Namespace), zap.Int("final_items", len(trace.FinalItems)))
	}
	h.DecisionRecorder.Record(&trace)
}

func buildTraceConstraints(c *types.PopConstraints) *audit.TraceConstraints {
	if c == nil {
		return nil
	}
	tc := &audit.TraceConstraints{}
	if len(c.IncludeTagsAny) > 0 {
		tc.IncludeTagsAny = append([]string(nil), c.IncludeTagsAny...)
	}
	if len(c.ExcludeItemIDs) > 0 {
		tc.ExcludeItemIDs = append([]string(nil), c.ExcludeItemIDs...)
	}
	if c.MinPrice != nil {
		tc.PriceBetween = append(tc.PriceBetween, *c.MinPrice)
	}
	if c.MaxPrice != nil {
		tc.PriceBetween = append(tc.PriceBetween, *c.MaxPrice)
	}
	if c.CreatedAfter != nil {
		tc.CreatedAfter = c.CreatedAfter.UTC().Format(time.RFC3339)
	}
	if len(tc.IncludeTagsAny) == 0 && len(tc.ExcludeItemIDs) == 0 && len(tc.PriceBetween) == 0 && tc.CreatedAfter == "" {
		return nil
	}
	return tc
}

func buildTraceCandidates(items []types.ScoredItem) []audit.TraceCandidate {
	if len(items) == 0 {
		return nil
	}
	copied := append([]types.ScoredItem(nil), items...)
	sort.SliceStable(copied, func(i, j int) bool { return copied[i].Score > copied[j].Score })
	out := make([]audit.TraceCandidate, len(copied))
	for i, cand := range copied {
		out[i] = audit.TraceCandidate{ItemID: cand.ItemID, Score: cand.Score}
	}
	return out
}

func buildTraceFinalItems(items []handlerstypes.ScoredItem, reasons map[string][]string) []audit.TraceFinalItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]audit.TraceFinalItem, 0, len(items))
	for _, item := range items {
		traceReasons := reasons[item.ItemID]
		if len(traceReasons) == 0 {
			traceReasons = item.Reasons
		}
		if len(traceReasons) > 0 {
			traceReasons = append([]string(nil), traceReasons...)
		}
		out = append(out, audit.TraceFinalItem{
			ItemID:  item.ItemID,
			Score:   item.Score,
			Reasons: traceReasons,
		})
	}
	return out
}

func buildTraceMMR(mmr map[string]algorithm.MMRExplain, caps map[string]algorithm.CapsExplain) []audit.TraceMMR {
	if len(mmr) == 0 {
		return nil
	}
	out := make([]audit.TraceMMR, 0, len(mmr))
	for itemID, info := range mmr {
		pick := info.Rank - 1
		if pick < 0 {
			pick = 0
		}
		capInfo := caps[itemID]
		out = append(out, audit.TraceMMR{
			PickIndex:      pick,
			ItemID:         itemID,
			MaxSimilarity:  info.MaxSimilarity,
			Relevance:      info.Relevance,
			Penalty:        info.Penalty,
			BrandCapHit:    capInfo.Brand != nil && capInfo.Brand.Applied,
			CategoryCapHit: capInfo.Category != nil && capInfo.Category.Applied,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PickIndex == out[j].PickIndex {
			return out[i].ItemID < out[j].ItemID
		}
		return out[i].PickIndex < out[j].PickIndex
	})
	return out
}

func buildTraceCaps(caps map[string]algorithm.CapsExplain) map[string]audit.TraceCap {
	if len(caps) == 0 {
		return nil
	}
	out := make(map[string]audit.TraceCap, len(caps))
	for itemID, cap := range caps {
		mapped := audit.TraceCap{}
		if cap.Brand != nil {
			limit, count := copyOptInt(cap.Brand.Limit), copyOptInt(cap.Brand.Count)
			mapped.Brand = &audit.TraceCapUsage{
				Applied: cap.Brand.Applied,
				Limit:   limit,
				Count:   count,
				Value:   cap.Brand.Value,
			}
		}
		if cap.Category != nil {
			limit, count := copyOptInt(cap.Category.Limit), copyOptInt(cap.Category.Count)
			mapped.Category = &audit.TraceCapUsage{
				Applied: cap.Category.Applied,
				Limit:   limit,
				Count:   count,
				Value:   cap.Category.Value,
			}
		}
		out[itemID] = mapped
	}
	return out
}

func buildBoostedList(boosted map[string]bool) []string {
	if len(boosted) == 0 {
		return nil
	}
	out := make([]string, 0, len(boosted))
	for itemID, ok := range boosted {
		if ok {
			out = append(out, itemID)
		}
	}
	sort.Strings(out)
	return out
}

func extractContextKeys(ctx map[string]any) []string {
	if len(ctx) == 0 {
		return nil
	}
	keys := make([]string, 0, len(ctx))
	for k := range ctx {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func copyOptInt(src *int) *int {
	if src == nil {
		return nil
	}
	v := *src
	return &v
}

func hashUser(namespace, userID, salt string) string {
	if namespace == "" || userID == "" || salt == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(namespace + ":" + userID + ":" + salt))
	return hex.EncodeToString(sum[:])
}
