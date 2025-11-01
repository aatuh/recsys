package explain

import "time"

// Request encapsulates ExplainLLM input parameters.
type Request struct {
	OrgID      string
	TargetType string
	TargetID   string
	Namespace  string
	Surface    string
	SegmentID  string
	From       time.Time
	To         time.Time
	Question   string
}

// Config holds LLM configuration and feature flags.
type Config struct {
	Enabled        bool
	Provider       string
	ModelPrimary   string
	ModelEscalate  string
	Timeout        time.Duration
	MaxTokens      int
	CacheTTL       time.Duration
	CircuitBreaker CircuitBreakerConfig
}

// FactsPack represents the compact facts sent to the LLM.
type FactsPack struct {
	Target struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	} `json:"target"`
	Window struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"window"`
	Context struct {
		Namespace string `json:"namespace"`
		Surface   string `json:"surface"`
		SegmentID string `json:"segment_id,omitempty"`
	} `json:"context"`
	Metrics     MetricsFacts      `json:"metrics"`
	RulesActive []RuleFact        `json:"rules_active"`
	Audit       []Evidence        `json:"audit"`
	Simulation  *SimulationFact   `json:"sim,omitempty"`
	Links       map[string]string `json:"links,omitempty"`
}

// MetricsFacts summarises key counters.
type MetricsFacts struct {
	Impressions int     `json:"impressions"`
	Clicks      int     `json:"clicks"`
	CTR         float64 `json:"ctr"`
	Errors      int     `json:"errors"`
}

// RuleFact captures an impacting rule.
type RuleFact struct {
	RuleID   string `json:"rule_id"`
	Action   string `json:"action"`
	Target   string `json:"target"`
	Priority int    `json:"priority"`
	TTL      string `json:"ttl,omitempty"`
}

// Evidence captures an audit evidence fact.
type Evidence struct {
	EvidenceID string `json:"evidence_id"`
	Stage      string `json:"stage"`
	Message    string `json:"message"`
	Count      int    `json:"count"`
}

// SimulationFact summarises dry-run eligibility.
type SimulationFact struct {
	EligibleNow bool   `json:"eligible_now"`
	Why         string `json:"why"`
}

// Result represents the explain response.
type Result struct {
	Markdown string
	Facts    FactsPack
	Model    string
	CacheHit bool
	Warnings []string
}

// CircuitBreakerConfig tunes protective wraps around LLM providers.
type CircuitBreakerConfig struct {
	Enabled           bool
	FailureThreshold  int
	ResetAfter        time.Duration
	HalfOpenSuccesses int
}
