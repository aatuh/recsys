package endpoints

// API version prefix
const V1 = "/v1"

// Ingestion endpoints
const (
	ItemsUpsert  = V1 + "/items:upsert"
	UsersUpsert  = V1 + "/users:upsert"
	EventsBatch  = V1 + "/events:batch"
	ItemsDelete  = V1 + "/items:delete"
	UsersDelete  = V1 + "/users:delete"
	EventsDelete = V1 + "/events:delete"
)

// Recommendation endpoints
const (
	Recommendations       = V1 + "/recommendations"
	ItemsSimilar          = V1 + "/items/{item_id}/similar"
	BanditRecommendations = V1 + "/bandit/recommendations"
)

// Bandit endpoints
const (
	BanditDecide         = V1 + "/bandit/decide"
	BanditPolicies       = V1 + "/bandit/policies"
	BanditPoliciesUpsert = V1 + "/bandit/policies:upsert"
	BanditReward         = V1 + "/bandit/reward"
)

// Configuration endpoints
const (
	EventTypes            = V1 + "/event-types"
	EventTypesUpsert      = V1 + "/event-types:upsert"
	Segments              = V1 + "/segments"
	SegmentsUpsert        = V1 + "/segments:upsert"
	SegmentsDelete        = V1 + "/segments:delete"
	SegmentProfiles       = V1 + "/segment-profiles"
	SegmentProfilesUpsert = V1 + "/segment-profiles:upsert"
	SegmentProfilesDelete = V1 + "/segment-profiles:delete"
	SegmentDryRun         = V1 + "/segments:dry-run"
)

// Data management endpoints
const (
	ItemsList  = V1 + "/items"
	UsersList  = V1 + "/users"
	EventsList = V1 + "/events"
)

// Audit endpoints
const (
	AuditDecisions    = V1 + "/audit/decisions"
	AuditDecisionByID = V1 + "/audit/decisions/{decision_id}"
	AuditSearch       = V1 + "/audit/search"
)

// Admin rule engine endpoints
const (
	Rules       = V1 + "/admin/rules"
	RuleByID    = V1 + "/admin/rules/{rule_id}"
	RulesDryRun = V1 + "/admin/rules/dry-run"
	ManualOverrides      = V1 + "/admin/manual_overrides"
	ManualOverrideCancel = V1 + "/admin/manual_overrides/{override_id}/cancel"
)

// Explain endpoints
const (
	ExplainLLM = V1 + "/explain/llm"
)

// Health and documentation
const (
	Health = "/health"
	Docs   = "/docs"
)

// Helper functions for endpoints with path parameters
func AuditDecisionByIDPath(decisionID string) string {
	return "/v1/audit/decisions/" + decisionID
}

func ItemsSimilarPath(itemID string) string {
	return "/v1/items/" + itemID + "/similar"
}
