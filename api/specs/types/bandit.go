package types

type BanditPolicy struct {
	PolicyID          string  `json:"policy_id"`
	Name              string  `json:"name"`
	Active            bool    `json:"active"`
	BlendAlpha        float64 `json:"blend_alpha"`
	BlendBeta         float64 `json:"blend_beta"`
	BlendGamma        float64 `json:"blend_gamma"`
	MMRLambda         float64 `json:"mmr_lambda"`
	BrandCap          int     `json:"brand_cap"`
	CategoryCap       int     `json:"category_cap"`
	ProfileBoost      float64 `json:"profile_boost"`
	RuleExcludeEvents bool    `json:"rule_exclude_events"`
	HalfLifeDays      float64 `json:"half_life_days"`
	CoVisWindowDays   int     `json:"co_vis_window_days"`
	PopularityFanout  int     `json:"popularity_fanout"`
	Notes             string  `json:"notes,omitempty"`
}

type BanditPoliciesUpsertRequest struct {
	Namespace string         `json:"namespace"`
	Policies  []BanditPolicy `json:"policies"`
}

type BanditDecideRequest struct {
	Namespace          string            `json:"namespace"`
	Surface            string            `json:"surface"`
	Context            map[string]string `json:"context"`
	CandidatePolicyIDs []string          `json:"candidate_policy_ids,omitempty"`
	Algorithm          string            `json:"algorithm,omitempty"`
	RequestID          string            `json:"request_id,omitempty"`
}

type BanditDecideResponse struct {
	PolicyID   string            `json:"policy_id"`
	Algorithm  string            `json:"algorithm"`
	Surface    string            `json:"surface"`
	BucketKey  string            `json:"bucket_key"`
	Explore    bool              `json:"explore"`
	Explain    map[string]string `json:"explain"`
	Experiment string            `json:"experiment,omitempty"`
	Variant    string            `json:"variant,omitempty"`
}

type BanditRewardRequest struct {
	Namespace  string `json:"namespace"`
	Surface    string `json:"surface"`
	BucketKey  string `json:"bucket_key"`
	PolicyID   string `json:"policy_id"`
	Reward     bool   `json:"reward"`
	Algorithm  string `json:"algorithm,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	Experiment string `json:"experiment,omitempty"`
	Variant    string `json:"variant,omitempty"`
}

// Wrapper: decide policy, then run ranker with it.
type RecommendWithBanditRequest struct {
	RecommendRequest
	Surface            string            `json:"surface"`
	Context            map[string]string `json:"context"`
	CandidatePolicyIDs []string          `json:"candidate_policy_ids,omitempty"`
	Algorithm          string            `json:"algorithm,omitempty"`
	RequestID          string            `json:"request_id,omitempty"`
}

type RecommendWithBanditResponse struct {
	RecommendResponse
	ChosenPolicyID   string            `json:"chosen_policy_id"`
	Algorithm        string            `json:"algorithm"`
	BanditBucket     string            `json:"bandit_bucket"`
	Explore          bool              `json:"explore"`
	BanditExplain    map[string]string `json:"bandit_explain"`
	RequestID        string            `json:"request_id,omitempty"`
	BanditExperiment string            `json:"bandit_experiment,omitempty"`
	BanditVariant    string            `json:"bandit_variant,omitempty"`
}
