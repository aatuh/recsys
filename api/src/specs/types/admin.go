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
