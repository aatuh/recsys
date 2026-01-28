package handlers

import (
	"time"

	"recsys/internal/segments"
	"recsys/internal/types"

	"github.com/aatuh/recsys-algo/algorithm"
)

func buildSegmentContextData(req algorithm.Request, ctxValues map[string]any, traits map[string]any) map[string]any {
	userData := map[string]any{
		"id": req.UserID,
	}
	if traits != nil {
		userData["traits"] = traits
	}

	ctxData := map[string]any{}
	for k, v := range ctxValues {
		ctxData[k] = v
	}

	requestData := map[string]any{
		"namespace": req.Namespace,
		"k":         req.K,
	}
	if req.Surface != "" {
		requestData["surface"] = req.Surface
	}
	if req.Blend != nil {
		requestData["blend"] = map[string]any{
			"pop":        req.Blend.Pop,
			"cooc":       req.Blend.Cooc,
			"similarity": req.Blend.Similarity,
		}
	}

	return map[string]any{
		"user":    userData,
		"ctx":     ctxData,
		"request": requestData,
	}
}

func segmentMatches(seg *types.Segment, data map[string]any, now time.Time) (*types.SegmentRule, bool) {
	if seg == nil {
		return nil, false
	}
	if len(seg.Rules) == 0 {
		return nil, true
	}
	for i := range seg.Rules {
		rule := seg.Rules[i]
		if !rule.Enabled {
			continue
		}
		eval := segments.NewEvaluator(data, now)
		matched, err := eval.Match(rule.Rule)
		if err != nil {
			continue
		}
		if matched {
			return &seg.Rules[i], true
		}
	}
	return nil, false
}
