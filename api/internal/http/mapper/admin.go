package mapper

import (
	"encoding/json"
	"time"

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

// AuditLogResponse maps audit log entries to DTO.
func AuditLogResponse(log adminsvc.AuditLog) types.AuditLogResponse {
	entries := make([]types.AuditLogEntry, 0, len(log.Entries))
	for _, entry := range log.Entries {
		entries = append(entries, AuditLogEntry(entry))
	}
	resp := types.AuditLogResponse{
		TenantID: log.TenantID,
		Entries:  entries,
	}
	if !log.NextBefore.IsZero() {
		resp.NextBefore = log.NextBefore.UTC().Format(time.RFC3339Nano)
		resp.NextBeforeID = log.NextBeforeID
	}
	return resp
}

// AuditLogEntry maps an audit entry to DTO.
func AuditLogEntry(entry adminsvc.AuditEntry) types.AuditLogEntry {
	return types.AuditLogEntry{
		ID:         entry.ID,
		OccurredAt: entry.OccurredAt.UTC().Format(time.RFC3339Nano),
		TenantID:   entry.TenantID,
		ActorSub:   entry.ActorSub,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		RequestID:  entry.RequestID,
		IP:         ipToString(entry.IP),
		UserAgent:  entry.UserAgent,
		Before:     decodeJSON(entry.Before),
		After:      decodeJSON(entry.After),
		Extra:      decodeJSON(entry.Extra),
	}
}

func decodeJSON(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	var payload any
	_ = json.Unmarshal(raw, &payload)
	return payload
}

func ipToString(ipAddr any) string {
	if ipAddr == nil {
		return ""
	}
	switch v := ipAddr.(type) {
	case interface{ String() string }:
		return v.String()
	default:
		return ""
	}
}
