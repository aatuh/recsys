package algorithm

import (
	"strings"
	"time"

	"recsys/internal/rules"
	"recsys/internal/types"

	"github.com/google/uuid"
)

// Config holds algorithm configuration parameters
type Config struct {
	// Blend weights
	BlendAlpha float64 // Popularity weight
	BlendBeta  float64 // Co-visitation weight
	BlendGamma float64 // Embedding weight

	// Personalization
	ProfileBoost               float64 // Multiplier for tag overlap
	ProfileWindowDays          float64 // Days to look back for user profile
	ProfileTopNTags            int     // Max tags in user profile
	ProfileMinEventsForBoost   int     // Minimum recent events before full boost applies
	ProfileColdStartMultiplier float64 // Attenuation factor for sparse history
	ProfileStarterBlendWeight  float64 // Weight (0-1) given to starter presets when blending with sparse history

	// MMR and diversity
	MMRLambda   float64 // MMR balance (0=diversity, 1=relevance)
	BrandCap    int     // Max items per brand (0=disabled)
	CategoryCap int     // Max items per category (0=disabled)

	// Windows and constraints
	HalfLifeDays        float64 // Popularity half-life
	CoVisWindowDays     int     // Co-visitation window
	PurchasedWindowDays int     // Exclude purchased window
	RuleExcludeEvents   bool    // Whether to exclude purchased items
	ExcludeEventTypes   []int16 // Event types to exclude when filtering user history
	BrandTagPrefixes    []string
	CategoryTagPrefixes []string

	RulesEnabled bool

	// Popularity fanout
	PopularityFanout int // Fanout for popularity candidates

	// Session retriever controls
	SessionLookbackEvents   int     // Number of recent events to seed session retriever
	SessionLookaheadMinutes float64 // horizon minutes for follow-up events
}

// Request represents a recommendation request
type Request struct {
	OrgID              uuid.UUID
	UserID             string
	Namespace          string
	Surface            string
	SegmentID          string
	K                  int
	Constraints        *types.PopConstraints
	Blend              *BlendWeights
	IncludeReasons     bool
	ExplainLevel       ExplainLevel
	StarterProfile     map[string]float64
	StarterBlendWeight float64
	RecentEventCount   int
	InjectAnchors      bool
	AnchorItemIDs      []string
}

// BlendWeights represents the blending weights for different signals
type BlendWeights struct {
	Pop  float64 // Popularity weight
	Cooc float64 // Co-visitation weight
	ALS  float64 // Embedding weight
}

// ExplainLevel controls how much structured explanation data to return.
type ExplainLevel string

const (
	ExplainLevelTags    ExplainLevel = "tags"
	ExplainLevelNumeric ExplainLevel = "numeric"
	ExplainLevelFull    ExplainLevel = "full"
)

// NormalizeExplainLevel converts a raw string into a known ExplainLevel.
func NormalizeExplainLevel(raw string) ExplainLevel {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(ExplainLevelNumeric):
		return ExplainLevelNumeric
	case string(ExplainLevelFull):
		return ExplainLevelFull
	default:
		return ExplainLevelTags
	}
}

// Response represents the recommendation response
type Response struct {
	ModelVersion string
	Items        []ScoredItem
	SegmentID    string
	ProfileID    string
}

// ScoredItem represents an item with its score and reasons
type ScoredItem struct {
	ItemID  string
	Score   float64
	Reasons []string
	Explain *ExplainBlock
}

// CandidateData holds all the data needed for recommendation
type CandidateData struct {
	Candidates        []types.ScoredItem
	Tags              map[string]types.ItemTags
	CoocScores        map[string]float64
	EmbScores         map[string]float64
	UsedCooc          map[string]bool
	UsedEmb           map[string]bool
	Collaborative     map[string]bool
	ContentBased      map[string]bool
	SessionBased      map[string]bool
	Boosted           map[string]bool
	PopNorm           map[string]float64
	CoocNorm          map[string]float64
	EmbNorm           map[string]float64
	PopRaw            map[string]float64
	CoocRaw           map[string]float64
	EmbRaw            map[string]float64
	ProfileOverlap    map[string]float64
	ProfileMultiplier map[string]float64
	Anchors           []string
	AnchorsFetched    bool
	MMRInfo           map[string]MMRExplain
	CapsInfo          map[string]CapsExplain
}

// TraceData aggregates algorithm internals for audit logging.
type TraceData struct {
	K                  int
	CandidatesPre      []types.ScoredItem
	MMRInfo            map[string]MMRExplain
	CapsInfo           map[string]CapsExplain
	Anchors            []string
	Boosted            map[string]bool
	Reasons            map[string][]string
	IncludeReasons     bool
	ExplainLevel       ExplainLevel
	ModelVersion       string
	RuleMatches        []rules.Match
	RuleEffects        map[string]rules.ItemEffect
	RuleEvaluated      []uuid.UUID
	RulePinned         []rules.PinnedItem
	SourceMetrics      map[string]SourceMetric
	Policy             *PolicySummary
	StarterProfile     map[string]float64
	StarterBlendWeight float64
	RecentEventCount   int
}

// SourceMetric captures coverage and latency for a candidate source.
type SourceMetric struct {
	Count    int           `json:"count"`
	Duration time.Duration `json:"duration"`
}

// PolicySummary captures enforcement stats for constraints and rule actions.
type PolicySummary struct {
	TotalCandidates          int                 `json:"total_candidates"`
	ExplicitExcludeHits      int                 `json:"explicit_exclude_hits"`
	RecentEventExcludeHits   int                 `json:"recent_event_exclude_hits"`
	AfterExclusions          int                 `json:"after_exclusions"`
	ConstraintIncludeTags    []string            `json:"constraint_include_tags,omitempty"`
	ConstraintFilteredCount  int                 `json:"constraint_filtered_count"`
	ConstraintFilteredIDs    []string            `json:"constraint_filtered_ids,omitempty"`
	constraintFilteredLookup map[string]struct{} `json:"-"`
	AfterConstraintFilters   int                 `json:"after_constraint_filters"`
	RuleBlockCount           int                 `json:"rule_block_count"`
	RulePinCount             int                 `json:"rule_pin_count"`
	RuleBoostCount           int                 `json:"rule_boost_count"`
	RuleBoostInjected        int                 `json:"rule_boost_injected"`
	RuleBoostExposure        int                 `json:"rule_boost_exposure"`
	RulePinExposure          int                 `json:"rule_pin_exposure"`
	AfterRules               int                 `json:"after_rules"`
	FinalCount               int                 `json:"final_count"`
	ConstraintLeakCount      int                 `json:"constraint_leak_count"`
	ConstraintLeakIDs        []string            `json:"constraint_leak_ids,omitempty"`
}

// SimilarItemsRequest represents a request for similar items
type SimilarItemsRequest struct {
	OrgID     uuid.UUID
	ItemID    string
	Namespace string
	K         int
}

// SimilarItemsResponse represents the response for similar items
type SimilarItemsResponse struct {
	Items []ScoredItem
}

// BlendContribution captures the weighted contribution from each signal.
type BlendContribution struct {
	Pop  float64 `json:"pop,omitempty"`
	Cooc float64 `json:"cooc,omitempty"`
	Emb  float64 `json:"emb,omitempty"`
}

// BlendRaw contains raw, unnormalized signal values for full explanations.
type BlendRaw struct {
	Pop  float64 `json:"pop,omitempty"`
	Cooc float64 `json:"cooc,omitempty"`
	Emb  float64 `json:"emb,omitempty"`
}

// BlendExplain provides context for how signals combined into a blended score.
type BlendExplain struct {
	Alpha         float64           `json:"alpha,omitempty"`
	Beta          float64           `json:"beta,omitempty"`
	Gamma         float64           `json:"gamma,omitempty"`
	PopNorm       float64           `json:"pop_norm,omitempty"`
	CoocNorm      float64           `json:"cooc_norm,omitempty"`
	EmbNorm       float64           `json:"emb_norm,omitempty"`
	Contributions BlendContribution `json:"contrib"`
	Raw           *BlendRaw         `json:"raw,omitempty"`
}

// PersonalizationExplain captures personalization overlap and boost info.
type PersonalizationExplain struct {
	Overlap         float64                    `json:"overlap,omitempty"`
	BoostMultiplier float64                    `json:"boost_multiplier,omitempty"`
	Raw             *PersonalizationExplainRaw `json:"raw,omitempty"`
}

// PersonalizationExplainRaw exposes configuration values in full mode.
type PersonalizationExplainRaw struct {
	ProfileBoost float64 `json:"profile_boost,omitempty"`
}

// MMRExplain details the diversity penalty applied during MMR.
type MMRExplain struct {
	Lambda        float64 `json:"lambda,omitempty"`
	MaxSimilarity float64 `json:"max_sim,omitempty"`
	Penalty       float64 `json:"penalty,omitempty"`
	Relevance     float64 `json:"relevance,omitempty"`
	Rank          int     `json:"rank,omitempty"`
}

// CapUsage indicates whether brand/category caps affected an item.
type CapUsage struct {
	Applied bool   `json:"applied"`
	Limit   *int   `json:"limit,omitempty"`
	Count   *int   `json:"count,omitempty"`
	Value   string `json:"value,omitempty"`
}

// CapsExplain aggregates cap usage for brand and category dimensions.
type CapsExplain struct {
	Brand    *CapUsage `json:"brand,omitempty"`
	Category *CapUsage `json:"category,omitempty"`
}

// ExplainBlock is returned per item when structured explanations are requested.
type ExplainBlock struct {
	Blend           *BlendExplain           `json:"blend,omitempty"`
	Personalization *PersonalizationExplain `json:"personalization,omitempty"`
	MMR             *MMRExplain             `json:"mmr,omitempty"`
	Caps            *CapsExplain            `json:"caps,omitempty"`
	Anchors         []string                `json:"anchors,omitempty"`
}
