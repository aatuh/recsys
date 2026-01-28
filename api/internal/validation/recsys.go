package validation

import (
	"net/http"
	"strings"
	"time"

	"recsys/internal/services/recsysvc"
	"recsys/src/specs/types"
)

const (
	defaultSegment        = "default"
	defaultK              = 20
	maxK                  = 200
	defaultMaxAnchors     = 50
	maxAnchorIDs          = 50
	maxIncludeIDs         = 5000
	maxExcludeIDs         = 5000
	maxTags               = 200
	warningDefaultApplied = "DEFAULT_APPLIED"
	codeInvalidRequest    = "RECSYS_INVALID_REQUEST"
	codeUnprocessable     = "RECSYS_UNPROCESSABLE_ENTITY"
)

// NormalizeRecommendRequest validates and normalizes a recommend request.
func NormalizeRecommendRequest(dto *types.RecommendRequest) (recsysvc.RecommendRequest, []recsysvc.Warning, error) {
	var warnings []recsysvc.Warning
	if dto == nil {
		return recsysvc.RecommendRequest{}, warnings, newError(http.StatusBadRequest, codeInvalidRequest, "payload required")
	}

	surface := strings.TrimSpace(dto.Surface)
	if surface == "" {
		return recsysvc.RecommendRequest{}, warnings, newError(http.StatusBadRequest, codeInvalidRequest, "surface is required")
	}

	segment := strings.TrimSpace(dto.Segment)
	if segment == "" {
		segment = defaultSegment
		warnings = appendWarning(warnings, warningDefaultApplied, "segment defaulted to 'default'")
	}

	k := defaultK
	if dto.K != nil {
		k = *dto.K
	} else {
		warnings = appendWarning(warnings, warningDefaultApplied, "k defaulted to 20")
	}
	if k < 1 || k > maxK {
		return recsysvc.RecommendRequest{}, warnings, newError(http.StatusBadRequest, codeInvalidRequest, "k must be between 1 and 200")
	}

	user := recsysvc.UserRef{}
	if dto.User != nil {
		user.UserID = strings.TrimSpace(dto.User.UserID)
		user.AnonymousID = strings.TrimSpace(dto.User.AnonymousID)
		user.SessionID = strings.TrimSpace(dto.User.SessionID)
	}
	if user.UserID == "" && user.AnonymousID == "" && user.SessionID == "" {
		return recsysvc.RecommendRequest{}, warnings, newError(http.StatusUnprocessableEntity, codeUnprocessable, "user_id or anonymous_id or session_id is required")
	}

	ctx, err := normalizeContext(dto.Context)
	if err != nil {
		return recsysvc.RecommendRequest{}, warnings, err
	}

	anchors, err := normalizeAnchors(dto.Anchors, &warnings)
	if err != nil {
		return recsysvc.RecommendRequest{}, warnings, err
	}

	candidates, err := normalizeCandidates(dto.Candidates)
	if err != nil {
		return recsysvc.RecommendRequest{}, warnings, err
	}

	constraints, err := normalizeConstraints(dto.Constraints)
	if err != nil {
		return recsysvc.RecommendRequest{}, warnings, err
	}

	weights, err := normalizeWeights(dto.Weights)
	if err != nil {
		return recsysvc.RecommendRequest{}, warnings, err
	}

	options, err := normalizeOptions(dto.Options, &warnings)
	if err != nil {
		return recsysvc.RecommendRequest{}, warnings, err
	}

	experiment := normalizeExperiment(dto.Experiment)

	return recsysvc.RecommendRequest{
		Surface:     surface,
		Segment:     segment,
		K:           k,
		User:        user,
		Context:     ctx,
		Anchors:     anchors,
		Candidates:  candidates,
		Constraints: constraints,
		Weights:     weights,
		Options:     options,
		Experiment:  experiment,
	}, warnings, nil
}

// NormalizeSimilarRequest validates and normalizes a similar-items request.
func NormalizeSimilarRequest(dto *types.SimilarRequest) (recsysvc.SimilarRequest, []recsysvc.Warning, error) {
	var warnings []recsysvc.Warning
	if dto == nil {
		return recsysvc.SimilarRequest{}, warnings, newError(http.StatusBadRequest, codeInvalidRequest, "payload required")
	}

	surface := strings.TrimSpace(dto.Surface)
	if surface == "" {
		return recsysvc.SimilarRequest{}, warnings, newError(http.StatusBadRequest, codeInvalidRequest, "surface is required")
	}

	segment := strings.TrimSpace(dto.Segment)
	if segment == "" {
		segment = defaultSegment
		warnings = appendWarning(warnings, warningDefaultApplied, "segment defaulted to 'default'")
	}

	itemID := strings.TrimSpace(dto.ItemID)
	if itemID == "" {
		return recsysvc.SimilarRequest{}, warnings, newError(http.StatusBadRequest, codeInvalidRequest, "item_id is required")
	}

	k := defaultK
	if dto.K != nil {
		k = *dto.K
	} else {
		warnings = appendWarning(warnings, warningDefaultApplied, "k defaulted to 20")
	}
	if k < 1 || k > maxK {
		return recsysvc.SimilarRequest{}, warnings, newError(http.StatusBadRequest, codeInvalidRequest, "k must be between 1 and 200")
	}

	constraints, err := normalizeConstraints(dto.Constraints)
	if err != nil {
		return recsysvc.SimilarRequest{}, warnings, err
	}

	options, err := normalizeOptions(dto.Options, &warnings)
	if err != nil {
		return recsysvc.SimilarRequest{}, warnings, err
	}

	return recsysvc.SimilarRequest{
		Surface:     surface,
		Segment:     segment,
		ItemID:      itemID,
		K:           k,
		Constraints: constraints,
		Options:     options,
	}, warnings, nil
}

func normalizeContext(ctx *types.RequestContext) (*recsysvc.RequestContext, error) {
	if ctx == nil {
		return nil, nil
	}
	out := recsysvc.RequestContext{
		Locale:  strings.TrimSpace(ctx.Locale),
		Device:  strings.TrimSpace(ctx.Device),
		Country: strings.TrimSpace(ctx.Country),
	}
	now := strings.TrimSpace(ctx.Now)
	if now != "" {
		parsed, err := time.Parse(time.RFC3339, now)
		if err != nil {
			return nil, newError(http.StatusBadRequest, codeInvalidRequest, "context.now must be RFC3339")
		}
		out.Now = parsed.UTC().Format(time.RFC3339)
	}
	return &out, nil
}

func normalizeAnchors(a *types.Anchors, warnings *[]recsysvc.Warning) (*recsysvc.Anchors, error) {
	if a == nil {
		return nil, nil
	}
	ids := normalizeList(a.ItemIDs, false)
	if len(ids) > maxAnchorIDs {
		return nil, newError(http.StatusBadRequest, codeInvalidRequest, "anchors.item_ids exceeds limit")
	}
	maxAnchors := defaultMaxAnchors
	if a.MaxAnchors != nil {
		maxAnchors = *a.MaxAnchors
	} else if warnings != nil {
		*warnings = appendWarning(*warnings, warningDefaultApplied, "anchors.max_anchors defaulted to 50")
	}
	if maxAnchors < len(ids) {
		return nil, newError(http.StatusUnprocessableEntity, codeUnprocessable, "anchors.max_anchors must be >= number of anchor items")
	}
	if maxAnchors <= 0 {
		return nil, newError(http.StatusBadRequest, codeInvalidRequest, "anchors.max_anchors must be positive")
	}
	return &recsysvc.Anchors{ItemIDs: ids, MaxAnchors: maxAnchors}, nil
}

func normalizeCandidates(c *types.Candidates) (*recsysvc.Candidates, error) {
	if c == nil {
		return nil, nil
	}
	includeIDs := normalizeList(c.IncludeIDs, false)
	excludeIDs := normalizeList(c.ExcludeIDs, false)
	if len(includeIDs) > maxIncludeIDs {
		return nil, newError(http.StatusBadRequest, codeInvalidRequest, "candidates.include_ids exceeds limit")
	}
	if len(excludeIDs) > maxExcludeIDs {
		return nil, newError(http.StatusBadRequest, codeInvalidRequest, "candidates.exclude_ids exceeds limit")
	}
	if len(includeIDs) == 0 && len(excludeIDs) == 0 {
		return nil, nil
	}
	return &recsysvc.Candidates{IncludeIDs: includeIDs, ExcludeIDs: excludeIDs}, nil
}

func normalizeConstraints(c *types.Constraints) (*recsysvc.Constraints, error) {
	if c == nil {
		return nil, nil
	}
	required := normalizeList(c.RequiredTags, true)
	forbidden := normalizeList(c.ForbiddenTags, true)
	if len(required) > maxTags || len(forbidden) > maxTags {
		return nil, newError(http.StatusBadRequest, codeInvalidRequest, "constraints tags exceed limit")
	}
	var maxPerTag map[string]int
	if len(c.MaxPerTag) > 0 {
		maxPerTag = make(map[string]int, len(c.MaxPerTag))
		for k, v := range c.MaxPerTag {
			key := strings.ToLower(strings.TrimSpace(k))
			if key == "" {
				continue
			}
			if v < 0 {
				return nil, newError(http.StatusUnprocessableEntity, codeUnprocessable, "constraints.max_per_tag values must be non-negative")
			}
			maxPerTag[key] = v
		}
	}
	if len(required) == 0 && len(forbidden) == 0 && len(maxPerTag) == 0 {
		return nil, nil
	}
	return &recsysvc.Constraints{RequiredTags: required, ForbiddenTags: forbidden, MaxPerTag: maxPerTag}, nil
}

func normalizeWeights(w *types.Weights) (*recsysvc.Weights, error) {
	if w == nil {
		return nil, nil
	}
	if w.Pop < 0 || w.Cooc < 0 || w.Emb < 0 {
		return nil, newError(http.StatusUnprocessableEntity, codeUnprocessable, "weights must be non-negative")
	}
	return &recsysvc.Weights{Pop: w.Pop, Cooc: w.Cooc, Emb: w.Emb}, nil
}

func normalizeOptions(o *types.Options, warnings *[]recsysvc.Warning) (recsysvc.Options, error) {
	opts := recsysvc.Options{IncludeReasons: false, Explain: "none", IncludeTrace: false, Seed: 0}
	if o == nil {
		if warnings != nil {
			*warnings = appendWarning(*warnings, warningDefaultApplied, "options defaulted")
		}
		return opts, nil
	}
	if o.IncludeReasons != nil {
		opts.IncludeReasons = *o.IncludeReasons
	}
	if o.IncludeTrace != nil {
		opts.IncludeTrace = *o.IncludeTrace
	}
	if o.Seed != nil {
		opts.Seed = *o.Seed
	}
	if strings.TrimSpace(o.Explain) != "" {
		opts.Explain = strings.ToLower(strings.TrimSpace(o.Explain))
	}
	if opts.Explain != "none" && opts.Explain != "summary" && opts.Explain != "full" {
		return recsysvc.Options{}, newError(http.StatusUnprocessableEntity, codeUnprocessable, "options.explain must be one of none, summary, full")
	}
	return opts, nil
}

func normalizeExperiment(e *types.Experiment) *recsysvc.Experiment {
	if e == nil {
		return nil
	}
	id := strings.TrimSpace(e.ID)
	variant := strings.TrimSpace(e.Variant)
	if id == "" && variant == "" {
		return nil
	}
	return &recsysvc.Experiment{ID: id, Variant: variant}
}

func normalizeList(values []string, lower bool) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		item := strings.TrimSpace(v)
		if lower {
			item = strings.ToLower(item)
		}
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func appendWarning(list []recsysvc.Warning, code, detail string) []recsysvc.Warning {
	if code == "" && detail == "" {
		return list
	}
	return append(list, recsysvc.Warning{Code: code, Detail: detail})
}
