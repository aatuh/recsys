package recsysvc

import "encoding/json"

// TenantConfig captures cached tenant configuration and versions.
type TenantConfig struct {
	TenantID string
	Surface  string
	Version  string
	Weights  *Weights
	Flags    map[string]bool
	Algo     string
}

// TenantRules captures cached tenant rules and versions.
type TenantRules struct {
	TenantID string
	Surface  string
	Version  string
	Raw      json.RawMessage
}
