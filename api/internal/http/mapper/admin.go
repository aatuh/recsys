package mapper

import (
	"encoding/json"

	"github.com/aatuh/recsys-suite/api/internal/services/adminsvc"
	"github.com/aatuh/recsys-suite/api/src/specs/types"
)

// TenantConfigResponse maps admin config to DTO.
func TenantConfigResponse(cfg adminsvc.TenantConfig) types.TenantConfigResponse {
	var payload any
	if len(cfg.Raw) > 0 {
		_ = json.Unmarshal(cfg.Raw, &payload)
	}
	return types.TenantConfigResponse{
		TenantID:      cfg.TenantID,
		ConfigVersion: cfg.Version,
		Config:        payload,
	}
}

// TenantRulesResponse maps admin rules to DTO.
func TenantRulesResponse(rules adminsvc.TenantRules) types.TenantRulesResponse {
	var payload any
	if len(rules.Raw) > 0 {
		_ = json.Unmarshal(rules.Raw, &payload)
	}
	return types.TenantRulesResponse{
		TenantID:     rules.TenantID,
		RulesVersion: rules.Version,
		Rules:        payload,
	}
}

// CacheInvalidateResponse maps cache invalidation result to DTO.
func CacheInvalidateResponse(result adminsvc.CacheInvalidateResult) types.CacheInvalidateResponse {
	return types.CacheInvalidateResponse{
		TenantID:    result.TenantID,
		Targets:     append([]string(nil), result.Targets...),
		Surface:     result.Surface,
		Status:      result.Status,
		Invalidated: result.Invalidated,
	}
}
