package algorithm

import (
	"recsys/internal/bandit"
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
	ProfileBoost      float64 // Multiplier for tag overlap
	ProfileWindowDays float64 // Days to look back for user profile
	ProfileTopNTags   int     // Max tags in user profile

	// MMR and diversity
	MMRLambda   float64 // MMR balance (0=diversity, 1=relevance)
	BrandCap    int     // Max items per brand (0=disabled)
	CategoryCap int     // Max items per category (0=disabled)

	// Windows and constraints
	HalfLifeDays         float64 // Popularity half-life
	CoVisWindowDays      int     // Co-visitation window
	PurchasedWindowDays  int     // Exclude purchased window
	RuleExcludePurchased bool    // Whether to exclude purchased items

	// Popularity fanout
	PopularityFanout int // Fanout for popularity candidates

	// Bandit algorithm
	BanditAlgo bandit.Algorithm
}

// Request represents a recommendation request
type Request struct {
	OrgID          uuid.UUID
	UserID         string
	Namespace      string
	K              int
	Constraints    *types.PopConstraints
	Blend          *BlendWeights
	IncludeReasons bool
}

// BlendWeights represents the blending weights for different signals
type BlendWeights struct {
	Pop  float64 // Popularity weight
	Cooc float64 // Co-visitation weight
	ALS  float64 // Embedding weight
}

// Response represents the recommendation response
type Response struct {
	ModelVersion string
	Items        []ScoredItem
}

// ScoredItem represents an item with its score and reasons
type ScoredItem struct {
	ItemID  string
	Score   float64
	Reasons []string
}

// CandidateData holds all the data needed for recommendation
type CandidateData struct {
	Candidates []types.ScoredItem
	Meta       map[string]types.ItemMeta
	CoocScores map[string]float64
	EmbScores  map[string]float64
	UsedCooc   map[string]bool
	UsedEmb    map[string]bool
	Boosted    map[string]bool
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
