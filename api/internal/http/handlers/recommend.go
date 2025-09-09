package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"recsys/internal/http/common"
	"recsys/internal/http/store"
	"recsys/internal/http/types"

	"github.com/go-chi/chi/v5"
)

// Recommend godoc
// @Summary      Get recommendations for a user
// @Tags         ranking
// @Accept       json
// @Produce      json
// @Param        payload  body  types.RecommendRequest  true  "Recommend request"
// @Success      200      {object}  types.RecommendResponse
// @Failure      400      {object}  common.APIError
// @Router       /v1/recommendations [post]
func (h *Handler) Recommend(w http.ResponseWriter, r *http.Request) {
	var req types.RecommendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.HttpError(w, r, err, http.StatusBadRequest)
		return
	}
	// Blended scoring (falls back to popularity-only if β=γ=0).
	h.scoreByBlend(w, r, req)
}

// ItemSimilar godoc
// @Summary      Get similar items
// @Tags         ranking
// @Produce      json
// @Param        item_id  path  string  true  "Item ID"
// @Param        namespace query string false "Namespace"  default(default)
// @Param        k        query int     false "Top-K"  default(20)
// @Success      200      {array}  types.ScoredItem
// @Failure      400      {object} common.APIError
// @Router       /v1/items/{item_id}/similar [get]
func (h *Handler) ItemSimilar(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "item_id")
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		ns = "default"
	}
	k := 20
	if s := r.URL.Query().Get("k"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			k = v
		}
	}
	orgID := h.defaultOrgFromHeader(r)

	//Prefer embedding neighbors (if both anchor and neighbors have embeddings).
	// If none found, fall back to co-visitation within the configured window.
	if emb, err := h.Store.SimilarByEmbeddingTopK(
		r.Context(), orgID, ns, itemID, k,
	); err == nil && len(emb) > 0 {
		_ = json.NewEncoder(w).Encode(emb)
		return
	}

	// Use COVIS_WINDOW_DAYS from handler config. Fall back to 30 days.
	days := h.CoVisWindowDays
	if days <= 0 {
		days = 30
	}
	window := time.Duration(days*24.0) * time.Hour
	since := time.Now().UTC().Add(-window)

	out, err := h.Store.CooccurrenceTopKWithin(
		r.Context(), orgID, ns, itemID, k, since,
	)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}
	resp := make([]types.ScoredItem, 0, len(out))
	for _, it := range out {
		resp = append(resp, types.ScoredItem{
			ItemID:  it.ItemID,
			Score:   it.Score,
			Reasons: []string{"co_visitation"},
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// scoreByBlend blends three signals over the popularity candidate pool:
//
//	final = α*pop_norm + β*cooc_norm + γ*emb_norm
//
// It falls back to popularity-only when β=γ=0. Reasons are included
// when requested. Personalization boost and MMR/caps are
// applied on top of the blended score.
func (h *Handler) scoreByBlend(
	w http.ResponseWriter,
	r *http.Request,
	req types.RecommendRequest,
) {
	orgID := h.defaultOrgFromHeader(r)

	k := req.K
	if k <= 0 {
		k = 20
	}

	// Popularity constraints mirrored from the existing scorer.
	var pc *store.PopConstraints
	if req.Constraints != nil {
		pc = &store.PopConstraints{}
		if len(req.Constraints.PriceBetween) >= 1 {
			v := req.Constraints.PriceBetween[0]
			pc.MinPrice = &v
		}
		if len(req.Constraints.PriceBetween) >= 2 {
			v := req.Constraints.PriceBetween[1]
			pc.MaxPrice = &v
		}
		if req.Constraints.CreatedAfterISO != "" {
			ts, err := time.Parse(time.RFC3339, req.Constraints.CreatedAfterISO)
			if err != nil {
				common.BadRequest(
					w, r, "invalid_created_after",
					"created_after must be RFC3339", nil,
				)
				return
			}
			pc.CreatedAfter = &ts
		}
		pc.IncludeTagsAny = req.Constraints.IncludeTagsAny
		pc.ExcludeItemIDs = req.Constraints.ExcludeItemIDs
	}

	// Fanout for candidate pool.
	fetchK := h.PopularityFanout
	if fetchK <= 0 || fetchK < k {
		fetchK = k
	}

	// Base: recent popularity candidates (already availability-aware).
	top, err := h.Store.PopularityTopK(
		r.Context(), orgID, req.Namespace, h.HalfLifeDays, fetchK, pc,
	)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	// Initial exclude set (request and optional "exclude purchased" rule).
	exclude := map[string]struct{}{}
	if req.Constraints != nil {
		for _, id := range req.Constraints.ExcludeItemIDs {
			exclude[id] = struct{}{}
		}
	}
	if h.RuleExcludePurchased && req.UserID != "" {
		lookback := time.Duration(h.PurchasedWindowDays*24.0) * time.Hour
		since := time.Now().UTC().Add(-lookback)
		bought, err := h.Store.ListUserPurchasedSince(
			r.Context(), orgID, req.Namespace, req.UserID, since,
		)
		if err != nil {
			common.HttpError(w, r, err, http.StatusInternalServerError)
			return
		}
		for _, id := range bought {
			exclude[id] = struct{}{}
		}
	}

	// Build candidate slice and a quick lookup set.
	candidates := make([]store.ScoredItem, 0, len(top))
	candSet := make(map[string]struct{}, len(top))
	maxPop := 0.0
	for _, it := range top {
		if _, skip := exclude[it.ItemID]; skip {
			continue
		}
		candidates = append(candidates, it)
		candSet[it.ItemID] = struct{}{}
		if it.Score > maxPop {
			maxPop = it.Score
		}
	}

	// Load metadata for caps/MMR and personalization.
	idList := make([]string, 0, len(candidates))
	for _, it := range candidates {
		idList = append(idList, it.ItemID)
	}
	meta, err := h.Store.ListItemsMeta(r.Context(), orgID, req.Namespace, idList)
	if err != nil {
		common.HttpError(w, r, err, http.StatusInternalServerError)
		return
	}

	// Resolve blend weights: request overrides defaults.
	alpha := h.BlendAlpha
	beta := h.BlendBeta
	gamma := h.BlendGamma
	if req.Blend != nil {
		alpha = math.Max(0, req.Blend.Pop)
		beta = math.Max(0, req.Blend.Cooc)
		gamma = math.Max(0, req.Blend.ALS)
	}
	if alpha == 0 && beta == 0 && gamma == 0 {
		alpha = 1 // safety: popularity-only
	}

	// If β or γ are non-zero, gather per-user co-vis/embedding signals
	// relative to the user's recent anchors, but only for candidate ids.
	cooc := map[string]float64{}
	emb := map[string]float64{}
	usedCooc := map[string]bool{}
	usedEmb := map[string]bool{}
	maxCooc := 0.0
	maxEmb := 0.0

	if req.UserID != "" && (beta > 0 || gamma > 0) {
		// Anchor window for co-vis; embedding neighbors do not need a time
		// window but still use recent anchors.
		days := h.CoVisWindowDays
		if days <= 0 {
			days = 30
		}
		since := time.Now().UTC().Add(-time.Duration(days*24.0) * time.Hour)

		// Limit number of recent anchors for performance.
		anchors, _ := h.Store.ListUserRecentItemIDs(
			r.Context(), orgID, req.Namespace, req.UserID, since, 10,
		)

		// Co-vis scores: max over anchors for each candidate.
		if beta > 0 && len(anchors) > 0 {
			for _, a := range anchors {
				neigh, err := h.Store.CooccurrenceTopKWithin(
					r.Context(), orgID, req.Namespace, a, 200, since,
				)
				if err != nil {
					continue
				}
				for _, n := range neigh {
					if _, ok := candSet[n.ItemID]; !ok {
						continue
					}
					if n.Score > cooc[n.ItemID] {
						cooc[n.ItemID] = n.Score
						usedCooc[n.ItemID] = true
						if n.Score > maxCooc {
							maxCooc = n.Score
						}
					}
				}
			}
		}

		// Embedding scores: max over anchors for each candidate.
		if gamma > 0 && len(anchors) > 0 {
			for _, a := range anchors {
				neigh, err := h.Store.SimilarByEmbeddingTopK(
					r.Context(), orgID, req.Namespace, a, 200,
				)
				if err != nil {
					continue
				}
				for _, n := range neigh {
					if _, ok := candSet[n.ItemID]; !ok {
						continue
					}
					if n.Score > emb[n.ItemID] {
						emb[n.ItemID] = n.Score
						usedEmb[n.ItemID] = true
						if n.Score > maxEmb {
							maxEmb = n.Score
						}
					}
				}
			}
		}
	}

	// Normalize and blend into final candidate scores.
	// If a channel's max is 0, its normalized value is 0 for all.
	for i := range candidates {
		id := candidates[i].ItemID

		popNorm := 0.0
		if maxPop > 0 {
			popNorm = candidates[i].Score / maxPop
		}
		coocNorm := 0.0
		if maxCooc > 0 {
			coocNorm = cooc[id] / maxCooc
		}
		embNorm := 0.0
		if maxEmb > 0 {
			embNorm = emb[id] / maxEmb
		}

		blended := alpha*popNorm + beta*coocNorm + gamma*embNorm
		candidates[i].Score = blended
	}

	// Light personalization boost on blended scores.
	usedPersonalization := false
	boosted := map[string]bool{}
	if req.UserID != "" && h.ProfileBoost > 0 {
		profile, err := h.Store.BuildUserTagProfile(
			r.Context(),
			orgID, req.Namespace, req.UserID,
			h.ProfileWindowDays,
			maxInt(h.ProfileTopNTags, 1),
		)
		if err == nil && len(profile) > 0 {
			usedPersonalization = true
			for i := range candidates {
				id := candidates[i].ItemID
				m := meta[id]
				overlap := 0.0
				for _, t := range m.Tags {
					if w, ok := profile[t]; ok {
						overlap += w
					}
				}
				if overlap > 0 {
					candidates[i].Score *= (1.0 + h.ProfileBoost*overlap)
					boosted[id] = true
				}
			}
		}
	}

	// Optional: MMR + caps.
	useMMR := h.MMRLambda > 0
	useCaps := h.BrandCap > 0 || h.CategoryCap > 0

	// Decide model_version for response.
	modelVersion := "blend_v1"
	if beta == 0 && gamma == 0 {
		modelVersion = "popularity_v1"
	}

	if !useMMR && !useCaps {
		resp := types.RecommendResponse{ModelVersion: modelVersion}
		for i, it := range candidates {
			if i >= k {
				break
			}
			rs := []string{}
			if alpha > 0 {
				rs = append(rs, "recent_popularity")
			}
			if usedCooc[it.ItemID] && beta > 0 {
				rs = append(rs, "co_visitation")
			}
			if usedEmb[it.ItemID] && gamma > 0 {
				rs = append(rs, "embedding_similarity")
			}
			if usedPersonalization && boosted[it.ItemID] && req.IncludeReasons {
				rs = append(rs, "personalization")
			}
			resp.Items = append(resp.Items, types.ScoredItem{
				ItemID:  it.ItemID,
				Score:   it.Score,
				Reasons: reasons(req.IncludeReasons, rs...),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// Re-rank with MMR + caps.
	reranked := mmrReRank(
		candidates, meta, k, h.MMRLambda, h.BrandCap, h.CategoryCap,
	)

	resp := types.RecommendResponse{ModelVersion: modelVersion}
	for _, it := range reranked {
		rs := []string{}
		if alpha > 0 {
			rs = append(rs, "recent_popularity")
		}
		if usedCooc[it.ItemID] && beta > 0 {
			rs = append(rs, "co_visitation")
		}
		if usedEmb[it.ItemID] && gamma > 0 {
			rs = append(rs, "embedding_similarity")
		}
		rs = append(rs, "diversity")
		if usedPersonalization && boosted[it.ItemID] && req.IncludeReasons {
			rs = append(rs, "personalization")
		}
		resp.Items = append(resp.Items, types.ScoredItem{
			ItemID:  it.ItemID,
			Score:   it.Score,
			Reasons: reasons(req.IncludeReasons, rs...),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// reasons returns the provided reason codes when include==true.
// It preserves order, de-duplicates, and does not whitelist codes.
// This allows new reasons like "personalization" without code changes.
func reasons(include bool, rs ...string) []string {
	if !include {
		return nil
	}
	seen := make(map[string]struct{}, len(rs))
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		if r == "" {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	return out
}

// mmrReRank performs a simple MMR on candidate items using tag overlap as
// a similarity proxy. It also enforces brand/category caps if provided.
//   - lambda in [0,1]. lambda=0 => no relevance term; lambda=1 => no
//     diversity term.
//   - brandCap/categoryCap 0 => disabled.
//   - If caps make further selection impossible, selection stops early.
func mmrReRank(
	candidates []store.ScoredItem,
	meta map[string]store.ItemMeta,
	k int,
	lambda float64,
	brandCap, categoryCap int,
) []store.ScoredItem {
	if k <= 0 {
		k = 1
	}
	out := make([]store.ScoredItem, 0, minInt(k, len(candidates)))
	if len(candidates) == 0 {
		return out
	}

	// Precompute normalized scores.
	maxScore := 0.0
	for _, c := range candidates {
		if c.Score > maxScore {
			maxScore = c.Score
		}
	}
	normScore := func(s float64) float64 {
		if maxScore <= 0 {
			return 0
		}
		return s / maxScore
	}

	// Prepare metadata into sets to speed up similarity calls.
	tagSet := map[string]map[string]struct{}{} // id -> tag set
	brandOf := map[string]string{}             // id -> brand value
	catOf := map[string]string{}               // id -> category value

	for id, m := range meta {
		sz := len(m.Tags)
		if sz == 0 {
			continue
		}
		set := make(map[string]struct{}, sz)
		for _, t := range m.Tags {
			lt := strings.ToLower(strings.TrimSpace(t))
			switch {
			case strings.HasPrefix(lt, "brand:"):
				brandOf[id] = strings.TrimSpace(lt[len("brand:"):])
			case strings.HasPrefix(lt, "category:"):
				catOf[id] = strings.TrimSpace(lt[len("category:"):])
			case strings.HasPrefix(lt, "cat:"):
				if _, ok := catOf[id]; !ok {
					catOf[id] = strings.TrimSpace(lt[len("cat:"):])
				}
			default:
				if lt != "" {
					set[lt] = struct{}{}
				}
			}
		}
		if len(set) > 0 {
			tagSet[id] = set
		}
	}

	brandCount := map[string]int{}
	catCount := map[string]int{}
	selected := map[string]struct{}{}

	remaining := append([]store.ScoredItem(nil), candidates...)

	// Greedy MMR with deterministic tie-break by initial order.
	for len(out) < k && len(remaining) > 0 {
		bestIdx := -1
		bestMMR := math.Inf(-1)

		for i, cand := range remaining {
			id := cand.ItemID

			// Enforce caps.
			if brandCap > 0 {
				if b := brandOf[id]; b != "" && brandCount[b] >= brandCap {
					continue
				}
			}
			if categoryCap > 0 {
				if c := catOf[id]; c != "" && catCount[c] >= categoryCap {
					continue
				}
			}

			// Diversity term: max similarity to already selected.
			maxSim := 0.0
			if len(selected) > 0 {
				for selID := range selected {
					s := jaccard(tagSet[id], tagSet[selID])
					if s > maxSim {
						maxSim = s
					}
				}
			}

			score := lambda*normScore(cand.Score) - (1.0-lambda)*maxSim
			if score > bestMMR {
				bestMMR = score
				bestIdx = i
			}
		}

		// If caps block all remaining items, stop selection early to
		// strictly enforce caps.
		if bestIdx == -1 {
			break
		}

		pick := remaining[bestIdx]
		out = append(out, pick)
		selected[pick.ItemID] = struct{}{}
		if b := brandOf[pick.ItemID]; b != "" {
			brandCount[b]++
		}
		if c := catOf[pick.ItemID]; c != "" {
			catCount[c]++
		}

		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}

	return out
}

// jaccard computes Jaccard similarity of two string sets.
func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	for k := range a {
		if _, ok := b[k]; ok {
			inter++
		}
	}
	union := len(a) + len(b) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
