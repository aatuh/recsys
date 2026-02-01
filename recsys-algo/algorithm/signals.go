package algorithm

import (
	"context"
	"errors"
	"math"
	"time"

	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"
)

// getBlendWeights returns the blend weights to use.
func (e *Engine) getBlendWeights(req Request) BlendWeights {
	weights := BlendWeights{
		Pop:        math.Max(0, e.config.BlendAlpha),
		Cooc:       math.Max(0, e.config.BlendBeta),
		Similarity: math.Max(0, e.config.BlendGamma),
	}

	// Override with request weights if provided.
	if req.Blend != nil {
		weights.Pop = math.Max(0, req.Blend.Pop)
		weights.Cooc = math.Max(0, req.Blend.Cooc)
		weights.Similarity = math.Max(0, req.Blend.Similarity)
	}

	// Safety: if all weights are 0, use popularity-only.
	if weights.Pop == 0 && weights.Cooc == 0 && weights.Similarity == 0 {
		weights.Pop = 1
	}

	return weights
}

// gatherSignals collects recent anchors, co-visitation and embedding signals.
func (e *Engine) gatherSignals(
	ctx context.Context,
	candidates []recmodel.ScoredItem,
	req Request,
	weights BlendWeights,
	sources map[string]SourceSet,
	signalStatus map[Signal]SignalStatus,
) (*CandidateData, error) {
	data := &CandidateData{
		Candidates:        candidates,
		Sources:           sources,
		CoocScores:        make(map[string]float64),
		EmbScores:         make(map[string]float64),
		SimilaritySources: make(map[string][]Signal),
		Boosted:           make(map[string]bool),
		PopNorm:           make(map[string]float64),
		CoocNorm:          make(map[string]float64),
		SimilarityNorm:    make(map[string]float64),
		PopRaw:            make(map[string]float64),
		CoocRaw:           make(map[string]float64),
		SimilarityRaw:     make(map[string]float64),
		ProfileOverlap:    make(map[string]float64),
		ProfileMultiplier: make(map[string]float64),
		SignalStatus:      signalStatus,
		MMRInfo:           make(map[string]MMRExplain),
		CapsInfo:          make(map[string]CapsExplain),
	}
	if data.Sources == nil {
		data.Sources = make(map[string]SourceSet)
	}
	if data.SignalStatus == nil {
		data.SignalStatus = make(map[Signal]SignalStatus)
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
			if errors.Is(err, recmodel.ErrFeatureUnavailable) {
				if weights.Cooc > 0 {
					e.recordSignalStatus(data.SignalStatus, SignalCooc, SignalOutcomeUnavailable, false, err)
				}
				if weights.Similarity > 0 {
					e.recordSignalStatus(data.SignalStatus, SignalEmbedding, SignalOutcomeUnavailable, false, err)
				}
			} else {
				if weights.Cooc > 0 {
					e.recordSignalStatus(data.SignalStatus, SignalCooc, SignalOutcomeError, false, err)
				}
				if weights.Similarity > 0 {
					e.recordSignalStatus(data.SignalStatus, SignalEmbedding, SignalOutcomeError, false, err)
				}
			}
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
	if weights.Cooc == 0 && weights.Similarity == 0 {
		return data, nil
	}

	// Co-vis signals.
	if weights.Cooc > 0 {
		if err := e.gatherCoVisitationSignals(ctx, data, req, data.Anchors, candSet); err != nil {
			return nil, err
		}
	}

	// Embedding signals.
	if weights.Similarity > 0 {
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
	store, ok := e.store.(recmodel.HistoryStore)
	if !ok {
		return nil, recmodel.ErrFeatureUnavailable
	}
	days := e.config.CoVisWindowDays
	if days <= 0 {
		days = defaultRecentAnchorsWindowDays
	}
	since := e.clock.Now().Add(-time.Duration(days*24.0) * time.Hour)

	return store.ListUserRecentItemIDs(
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
	store, ok := e.store.(recmodel.CooccurrenceStore)
	if !ok {
		e.recordSignalStatus(data.SignalStatus, SignalCooc, SignalOutcomeUnavailable, false, recmodel.ErrFeatureUnavailable)
		return nil
	}
	days := e.config.CoVisWindowDays
	if days <= 0 {
		days = defaultCovisWindowDays
	}
	since := e.clock.Now().Add(-time.Duration(days*24.0) * time.Hour)

	success := false
	failures := 0
	var lastErr error
	for _, anchor := range anchors {
		if err := ctx.Err(); err != nil {
			return err
		}
		// Gather co-visitation neighbors.
		neighbors, err := store.CooccurrenceTopKWithin(
			ctx,
			req.OrgID,
			req.Namespace,
			anchor,
			maxCovisNeighbors,
			since,
		)
		if err != nil {
			if errors.Is(err, recmodel.ErrFeatureUnavailable) {
				e.recordSignalStatus(data.SignalStatus, SignalCooc, SignalOutcomeUnavailable, false, err)
				return nil
			}
			failures++
			lastErr = err
			continue // Skip this anchor on error.
		}
		success = true

		// Update co-visitation scores.
		for _, neighbor := range neighbors {
			if err := ctx.Err(); err != nil {
				return err
			}
			if _, ok := candSet[neighbor.ItemID]; !ok {
				continue // Not in candidate set.
			}
			// Set score if higher than current score.
			if neighbor.Score > data.CoocScores[neighbor.ItemID] {
				data.CoocScores[neighbor.ItemID] = neighbor.Score
				addSource(data.Sources, neighbor.ItemID, SignalCooc)
			}
		}
	}

	if success {
		e.recordSignalStatus(data.SignalStatus, SignalCooc, SignalOutcomeSuccess, failures > 0, nil)
	} else if failures > 0 {
		e.recordSignalStatus(data.SignalStatus, SignalCooc, SignalOutcomeError, false, lastErr)
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
	store, ok := e.store.(recmodel.EmbeddingStore)
	if !ok {
		e.recordSignalStatus(data.SignalStatus, SignalEmbedding, SignalOutcomeUnavailable, false, recmodel.ErrFeatureUnavailable)
		return nil
	}
	success := false
	failures := 0
	var lastErr error
	for _, anchor := range anchors {
		if err := ctx.Err(); err != nil {
			return err
		}
		// Gather embedding neighbors.
		neighbors, err := store.SimilarByEmbeddingTopK(
			ctx,
			req.OrgID,
			req.Namespace,
			anchor,
			maxEmbNeighbors,
		)
		if err != nil {
			if errors.Is(err, recmodel.ErrFeatureUnavailable) {
				e.recordSignalStatus(data.SignalStatus, SignalEmbedding, SignalOutcomeUnavailable, false, err)
				return nil
			}
			failures++
			lastErr = err
			continue // Skip this anchor on error.
		}
		success = true

		// Update embedding scores.
		for _, neighbor := range neighbors {
			if err := ctx.Err(); err != nil {
				return err
			}
			if _, ok := candSet[neighbor.ItemID]; !ok {
				continue // Not in candidate set.
			}
			// Max aggregate: set score if higher than current score.
			if neighbor.Score > data.EmbScores[neighbor.ItemID] {
				data.EmbScores[neighbor.ItemID] = neighbor.Score
				addSource(data.Sources, neighbor.ItemID, SignalEmbedding)
			}
		}
	}

	if success {
		e.recordSignalStatus(data.SignalStatus, SignalEmbedding, SignalOutcomeSuccess, failures > 0, nil)
	} else if failures > 0 {
		e.recordSignalStatus(data.SignalStatus, SignalEmbedding, SignalOutcomeError, false, lastErr)
	}

	return nil
}

func addSource(sources map[string]SourceSet, itemID string, source Signal) {
	if itemID == "" {
		return
	}
	set := sources[itemID]
	if set == nil {
		set = make(SourceSet)
		sources[itemID] = set
	}
	set[source] = struct{}{}
}

func (e *Engine) recordSignalStatus(
	signalStatus map[Signal]SignalStatus,
	signal Signal,
	outcome SignalOutcome,
	partial bool,
	err error,
) {
	if signalStatus == nil {
		return
	}

	entry := SignalStatus{Available: outcome == SignalOutcomeSuccess, Partial: partial}
	if outcome != SignalOutcomeSuccess {
		if err != nil {
			entry.Error = err.Error()
		}
		if outcome == SignalOutcomeUnavailable {
			entry.Error = "unavailable"
		}
	}
	signalStatus[signal] = entry
	if e.signalObserver != nil {
		e.signalObserver.RecordSignal(signal, outcome)
	}
}
