package recsysvc

// RecommendRequest is the normalized domain request for recommendations.
type RecommendRequest struct {
	Surface     string
	Segment     string
	K           int
	User        UserRef
	Context     *RequestContext
	Anchors     *Anchors
	Candidates  *Candidates
	Constraints *Constraints
	Weights     *Weights
	Options     Options
	Experiment  *Experiment
	Algorithm   string
}

// SimilarRequest is the normalized domain request for similar items.
type SimilarRequest struct {
	Surface     string
	Segment     string
	ItemID      string
	K           int
	Constraints *Constraints
	Options     Options
	Algorithm   string
}

// UserRef carries user/session identifiers.
type UserRef struct {
	UserID      string
	AnonymousID string
	SessionID   string
}

// RequestContext captures request context metadata.
type RequestContext struct {
	Locale  string
	Device  string
	Country string
	Now     string
}

// Anchors contains optional anchor item IDs.
type Anchors struct {
	ItemIDs    []string
	MaxAnchors int
}

// Candidates optionally constrains candidate items.
type Candidates struct {
	IncludeIDs []string
	ExcludeIDs []string
}

// Constraints applies tag and diversity constraints.
type Constraints struct {
	RequiredTags  []string
	ForbiddenTags []string
	MaxPerTag     map[string]int
}

// Weights controls signal weights used in scoring.
type Weights struct {
	Pop  float64
	Cooc float64
	Emb  float64
}

// Options configures response details.
type Options struct {
	IncludeReasons bool
	Explain        string
	IncludeTrace   bool
	Seed           int64
}

// Experiment carries optional experiment metadata.
type Experiment struct {
	ID      string
	Variant string
}

// Item represents a ranked recommendation result.
type Item struct {
	ItemID  string
	Rank    int
	Score   float64
	Reasons []string
	Explain *ItemExplain
}

// ItemExplain provides optional explain metadata.
type ItemExplain struct {
	Signals map[string]float64
	Rules   []string
}

// Warning describes a non-fatal warning.
type Warning struct {
	Code   string
	Detail string
}
