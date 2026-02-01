package algorithm

import (
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/api/recsys-algo/rules"

	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"

	"github.com/google/uuid"
)

// Signal represents a scoring or retrieval signal in the pipeline.
type Signal string

const (
	SignalPop           Signal = "popularity"
	SignalCooc          Signal = "cooc"
	SignalEmbedding     Signal = "embedding"
	SignalCollaborative Signal = "collaborative"
	SignalContent       Signal = "content"
	SignalSession       Signal = "session"
)

// AlgorithmKind selects the baseline recommendation strategy.
type AlgorithmKind string

const (
	AlgorithmBlend      AlgorithmKind = "blend"
	AlgorithmPopularity AlgorithmKind = "popularity"
	AlgorithmCooc       AlgorithmKind = "cooc"
	AlgorithmImplicit   AlgorithmKind = "implicit"
)

// SourceSet tracks which signals contributed data for a candidate.
type SourceSet map[Signal]struct{}

// Config holds algorithm configuration parameters
type Config struct {
	// DefaultAlgorithm selects the default algorithm when a request does not specify one.
	DefaultAlgorithm AlgorithmKind
	// Version is an optional version label for the algorithm build.
	Version string

	// Blend weights
	BlendAlpha float64 // Popularity weight
	BlendBeta  float64 // Co-visitation weight
	BlendGamma float64 // Similarity weight (embedding/collab/content/session)

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
	PopularityFanout   int // Fanout for popularity candidates
	MaxK               int // Cap for requested K (0=disabled)
	MaxFanout          int // Cap for candidate fanout (0=disabled)
	MaxExcludeIDs      int // Cap for explicit exclude IDs (0=disabled)
	MaxAnchorsInjected int // Cap for injected anchors (0=disabled)

	// Session retriever controls
	SessionLookbackEvents   int     // Number of recent events to seed session retriever
	SessionLookaheadMinutes float64 // horizon minutes for follow-up events
}

// Request represents a recommendation request
type Request struct {
	OrgID                uuid.UUID
	UserID               string
	Namespace            string
	Surface              string
	SegmentID            string
	K                    int
	Algorithm            AlgorithmKind
	Constraints          *recmodel.PopConstraints
	Blend                *BlendWeights
	IncludeReasons       bool
	ExplainLevel         ExplainLevel
	StarterProfile       map[string]float64
	StarterBlendWeight   float64
	RecentEventCount     int
	InjectAnchors        bool
	AnchorItemIDs        []string
	PrefetchedCandidates []recmodel.ScoredItem
}

// BlendWeights represents the blending weights for different signals
type BlendWeights struct {
	Pop        float64 // Popularity weight
	Cooc       float64 // Co-visitation weight
	Similarity float64 // Similarity weight (embedding/collab/content/session)
}

const (
	ModelVersionPopularity = "popularity_v1"
	ModelVersionBlend      = "blend_v1"
)

// ModelVersionForWeights derives the model version label given blend weights.
func ModelVersionForWeights(weights BlendWeights) string {
	if weights.Cooc == 0 && weights.Similarity == 0 {
		return ModelVersionPopularity
	}
	return ModelVersionBlend
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

// NormalizeAlgorithm converts a raw algorithm selection into a supported value.
func NormalizeAlgorithm(raw AlgorithmKind) AlgorithmKind {
	switch strings.ToLower(strings.TrimSpace(string(raw))) {
	case string(AlgorithmPopularity):
		return AlgorithmPopularity
	case string(AlgorithmCooc):
		return AlgorithmCooc
	case string(AlgorithmImplicit):
		return AlgorithmImplicit
	case string(AlgorithmBlend):
		fallthrough
	default:
		return AlgorithmBlend
	}
}

// IsSupportedAlgorithm reports whether the value is a known algorithm kind.
func IsSupportedAlgorithm(raw AlgorithmKind) bool {
	switch strings.ToLower(strings.TrimSpace(string(raw))) {
	case string(AlgorithmPopularity), string(AlgorithmCooc), string(AlgorithmImplicit), string(AlgorithmBlend):
		return true
	default:
		return false
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
	Candidates        []recmodel.ScoredItem
	Tags              map[string]recmodel.ItemTags
	Sources           map[string]SourceSet
	PopScores         map[string]float64
	CoocScores        map[string]float64
	EmbScores         map[string]float64
	CollabScores      map[string]float64
	ContentScores     map[string]float64
	SessionScores     map[string]float64
	SimilaritySources map[string][]Signal
	Boosted           map[string]bool
	PopNorm           map[string]float64
	CoocNorm          map[string]float64
	SimilarityNorm    map[string]float64
	PopRaw            map[string]float64
	CoocRaw           map[string]float64
	SimilarityRaw     map[string]float64
	ProfileOverlap    map[string]float64
	ProfileMultiplier map[string]float64
	Anchors           []string
	AnchorsFetched    bool
	SignalStatus      map[Signal]SignalStatus
	MMRInfo           map[string]MMRExplain
	CapsInfo          map[string]CapsExplain
}

// TraceData aggregates algorithm internals for audit logging.
type TraceData struct {
	K                  int
	CandidatesPre      []recmodel.ScoredItem
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
	SignalStatus       map[Signal]SignalStatus
	Policy             *PolicySummary
	StarterProfile     map[string]float64
	StarterBlendWeight float64
	RecentEventCount   int
	ManualOverrideHits []rules.OverrideHit
}

// SourceMetric captures coverage and latency for a candidate source.
type SourceMetric struct {
	Count    int           `json:"count"`
	Duration time.Duration `json:"duration"`
}

// SignalStatus captures availability for a signal during a recommendation pass.
type SignalStatus struct {
	Available bool   `json:"available"`
	Partial   bool   `json:"partial,omitempty"`
	Error     string `json:"error,omitempty"`
}

// PolicySummary captures enforcement stats for constraints and rule actions.
type PolicySummary struct {
	TotalCandidates           int                 `json:"total_candidates"`
	ExplicitExcludeHits       int                 `json:"explicit_exclude_hits"`
	RecentEventExcludeHits    int                 `json:"recent_event_exclude_hits"`
	AfterExclusions           int                 `json:"after_exclusions"`
	ConstraintIncludeTags     []string            `json:"constraint_include_tags,omitempty"`
	ConstraintFilteredCount   int                 `json:"constraint_filtered_count"`
	ConstraintFilteredIDs     []string            `json:"constraint_filtered_ids,omitempty"`
	constraintFilteredLookup  map[string]struct{} `json:"-"`
	constraintFilteredReasons map[string]string   `json:"-"`
	AfterConstraintFilters    int                 `json:"after_constraint_filters"`
	RuleBlockCount            int                 `json:"rule_block_count"`
	RulePinCount              int                 `json:"rule_pin_count"`
	RuleBoostCount            int                 `json:"rule_boost_count"`
	RuleBoostInjected         int                 `json:"rule_boost_injected"`
	RuleBlockExposure         int                 `json:"rule_block_exposure"`
	RuleBoostExposure         int                 `json:"rule_boost_exposure"`
	RulePinExposure           int                 `json:"rule_pin_exposure"`
	AfterRules                int                 `json:"after_rules"`
	FinalCount                int                 `json:"final_count"`
	ConstraintLeakCount       int                 `json:"constraint_leak_count"`
	ConstraintLeakIDs         []string            `json:"constraint_leak_ids,omitempty"`
	ConstraintLeakByReason    map[string]int      `json:"constraint_leak_by_reason,omitempty"`
	RuleBlockExposureByRule   map[string]int      `json:"rule_block_exposure_by_rule,omitempty"`
}

// SimilarItemsRequest represents a request for similar items
type SimilarItemsRequest struct {
	OrgID     uuid.UUID
	ItemID    string
	Namespace string
	K         int
	Algorithm AlgorithmKind
}

// SimilarItemsResponse represents the response for similar items
type SimilarItemsResponse struct {
	Items []ScoredItem
}

// BlendContribution captures the weighted contribution from each signal.
type BlendContribution struct {
	Pop        float64 `json:"pop,omitempty"`
	Cooc       float64 `json:"cooc,omitempty"`
	Similarity float64 `json:"similarity,omitempty"`
}

// BlendRaw contains raw, unnormalized signal values for full explanations.
type BlendRaw struct {
	Pop        float64 `json:"pop,omitempty"`
	Cooc       float64 `json:"cooc,omitempty"`
	Similarity float64 `json:"similarity,omitempty"`
}

// BlendExplain provides context for how signals combined into a blended score.
type BlendExplain struct {
	Alpha          float64           `json:"alpha,omitempty"`
	Beta           float64           `json:"beta,omitempty"`
	Gamma          float64           `json:"gamma,omitempty"`
	PopNorm        float64           `json:"pop_norm,omitempty"`
	CoocNorm       float64           `json:"cooc_norm,omitempty"`
	SimilarityNorm float64           `json:"similarity_norm,omitempty"`
	Contributions  BlendContribution `json:"contrib"`
	Raw            *BlendRaw         `json:"raw,omitempty"`
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
