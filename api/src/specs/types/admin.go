package types

// TenantConfigResponse represents current tenant config.
type TenantConfigResponse struct {
	TenantID      string `json:"tenant_id"`
	ConfigVersion string `json:"config_version"`
	Config        any    `json:"config"`
}

// TenantRulesResponse represents current tenant rules.
type TenantRulesResponse struct {
	TenantID     string `json:"tenant_id"`
	RulesVersion string `json:"rules_version"`
	Rules        any    `json:"rules"`
}

// CacheInvalidateRequest describes a cache invalidation request.
type CacheInvalidateRequest struct {
	Targets []string `json:"targets"`
	Surface string   `json:"surface,omitempty"`
}

// CacheInvalidateResponse describes cache invalidation response.
type CacheInvalidateResponse struct {
	TenantID    string         `json:"tenant_id"`
	Targets     []string       `json:"targets"`
	Surface     string         `json:"surface,omitempty"`
	Status      string         `json:"status"`
	Invalidated map[string]int `json:"invalidated,omitempty"`
}

// AuditLogEntry represents an audit log row.
type AuditLogEntry struct {
	ID         int64  `json:"id"`
	OccurredAt string `json:"occurred_at"`
	TenantID   string `json:"tenant_id"`
	ActorSub   string `json:"actor_sub"`
	ActorType  string `json:"actor_type"`
	Action     string `json:"action"`
	EntityType string `json:"entity_type,omitempty"`
	EntityID   string `json:"entity_id,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	IP         string `json:"ip,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
	Before     any    `json:"before_state,omitempty"`
	After      any    `json:"after_state,omitempty"`
	Extra      any    `json:"extra,omitempty"`
}

// AuditLogResponse returns paginated audit entries.
type AuditLogResponse struct {
	TenantID     string          `json:"tenant_id"`
	Entries      []AuditLogEntry `json:"entries"`
	NextBefore   string          `json:"next_before,omitempty"`
	NextBeforeID int64           `json:"next_before_id,omitempty"`
}
