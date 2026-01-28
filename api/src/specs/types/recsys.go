package types

// RecommendRequest represents the public recommend API request payload.
type RecommendRequest struct {
	Surface     string          `json:"surface"`
	Segment     string          `json:"segment,omitempty"`
	K           *int            `json:"k,omitempty"`
	User        *UserRef        `json:"user,omitempty"`
	Context     *RequestContext `json:"context,omitempty"`
	Anchors     *Anchors        `json:"anchors,omitempty"`
	Candidates  *Candidates     `json:"candidates,omitempty"`
	Constraints *Constraints    `json:"constraints,omitempty"`
	Weights     *Weights        `json:"weights,omitempty"`
	Options     *Options        `json:"options,omitempty"`
	Experiment  *Experiment     `json:"experiment,omitempty"`
}

// SimilarRequest represents the public similar-items API request payload.
type SimilarRequest struct {
	Surface     string       `json:"surface"`
	Segment     string       `json:"segment,omitempty"`
	ItemID      string       `json:"item_id"`
	K           *int         `json:"k,omitempty"`
	Constraints *Constraints `json:"constraints,omitempty"`
	Options     *Options     `json:"options,omitempty"`
}

// NormalizedRecommendRequest is returned by /v1/recommend/validate.
type NormalizedRecommendRequest struct {
	Surface     string             `json:"surface"`
	Segment     string             `json:"segment"`
	K           int                `json:"k"`
	User        *UserRef           `json:"user,omitempty"`
	Context     *RequestContext    `json:"context,omitempty"`
	Anchors     *AnchorsNormalized `json:"anchors,omitempty"`
	Candidates  *Candidates        `json:"candidates,omitempty"`
	Constraints *Constraints       `json:"constraints,omitempty"`
	Weights     *Weights           `json:"weights,omitempty"`
	Options     *OptionsNormalized `json:"options,omitempty"`
	Experiment  *Experiment        `json:"experiment,omitempty"`
}

// UserRef carries user/session identifiers.
type UserRef struct {
	UserID      string `json:"user_id,omitempty"`
	AnonymousID string `json:"anonymous_id,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
}

// RequestContext captures request context metadata.
type RequestContext struct {
	Locale  string `json:"locale,omitempty"`
	Device  string `json:"device,omitempty"`
	Country string `json:"country,omitempty"`
	Now     string `json:"now,omitempty"`
}

// Anchors contains optional anchor item IDs.
type Anchors struct {
	ItemIDs    []string `json:"item_ids,omitempty"`
	MaxAnchors *int     `json:"max_anchors,omitempty"`
}

// AnchorsNormalized contains normalized anchor data.
type AnchorsNormalized struct {
	ItemIDs    []string `json:"item_ids,omitempty"`
	MaxAnchors int      `json:"max_anchors"`
}

// Candidates optionally constrains candidate items.
type Candidates struct {
	IncludeIDs []string `json:"include_ids,omitempty"`
	ExcludeIDs []string `json:"exclude_ids,omitempty"`
}

// Constraints applies tag and diversity constraints.
type Constraints struct {
	RequiredTags  []string       `json:"required_tags,omitempty"`
	ForbiddenTags []string       `json:"forbidden_tags,omitempty"`
	MaxPerTag     map[string]int `json:"max_per_tag,omitempty"`
}

// Weights controls signal weights used in scoring.
type Weights struct {
	Pop  float64 `json:"pop,omitempty"`
	Cooc float64 `json:"cooc,omitempty"`
	Emb  float64 `json:"emb,omitempty"`
}

// Options configures response details.
type Options struct {
	IncludeReasons *bool  `json:"include_reasons,omitempty"`
	Explain        string `json:"explain,omitempty"`
	IncludeTrace   *bool  `json:"include_trace,omitempty"`
	Seed           *int64 `json:"seed,omitempty"`
}

// OptionsNormalized contains normalized options.
type OptionsNormalized struct {
	IncludeReasons bool   `json:"include_reasons"`
	Explain        string `json:"explain"`
	IncludeTrace   bool   `json:"include_trace"`
	Seed           int64  `json:"seed"`
}

// Experiment carries optional experiment metadata.
type Experiment struct {
	ID      string `json:"id,omitempty"`
	Variant string `json:"variant,omitempty"`
}

// RecommendItem represents an item in the recommend response.
type RecommendItem struct {
	ItemID  string                `json:"item_id"`
	Rank    int                   `json:"rank"`
	Score   float64               `json:"score"`
	Reasons []string              `json:"reasons,omitempty"`
	Explain *RecommendItemExplain `json:"explain,omitempty"`
}

// RecommendItemExplain provides optional explain metadata.
type RecommendItemExplain struct {
	Signals map[string]float64 `json:"signals,omitempty"`
	Rules   []string           `json:"rules,omitempty"`
}

// ResponseMeta provides response metadata.
type ResponseMeta struct {
	TenantID      string         `json:"tenant_id,omitempty"`
	Surface       string         `json:"surface,omitempty"`
	Segment       string         `json:"segment,omitempty"`
	AlgoVersion   string         `json:"algo_version,omitempty"`
	ConfigVersion string         `json:"config_version,omitempty"`
	RulesVersion  string         `json:"rules_version,omitempty"`
	RequestID     string         `json:"request_id,omitempty"`
	TimingsMS     map[string]int `json:"timings_ms,omitempty"`
	Counts        map[string]int `json:"counts,omitempty"`
}

// Warning describes a non-fatal warning.
type Warning struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

// RecommendResponse is returned by /v1/recommend and /v1/similar.
type RecommendResponse struct {
	Items    []RecommendItem `json:"items"`
	Meta     ResponseMeta    `json:"meta"`
	Warnings []Warning       `json:"warnings,omitempty"`
}

// ValidateResponse is returned by /v1/recommend/validate.
type ValidateResponse struct {
	NormalizedRequest NormalizedRecommendRequest `json:"normalized_request"`
	Warnings          []Warning                  `json:"warnings,omitempty"`
	Meta              ResponseMeta               `json:"meta"`
}
