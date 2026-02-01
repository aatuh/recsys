package store

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/aatuh/api-toolkit-contrib/adapters/txpostgres"
	"github.com/aatuh/api-toolkit/ports"
	"github.com/aatuh/recsys-suite/api/recsys-algo/rules"
	"github.com/google/uuid"
)

type rulePayload map[string]any

// RulesManagerStore loads tenant rules for recsys-algo's rules manager.
type RulesManagerStore struct {
	Pool ports.DatabasePool
}

func NewRulesManagerStore(pool ports.DatabasePool) *RulesManagerStore {
	return &RulesManagerStore{Pool: pool}
}

// ListActiveRulesForScope returns active rules for the requested scope.
func (s *RulesManagerStore) ListActiveRulesForScope(
	ctx context.Context,
	orgID uuid.UUID,
	namespace, surface, segmentID string,
	ts time.Time,
) ([]rules.Rule, error) {
	if s == nil || s.Pool == nil {
		return nil, nil
	}
	tenantID, err := resolveTenantID(ctx, s.Pool, orgID)
	if err != nil {
		return nil, err
	}
	if tenantID == uuid.Nil {
		return nil, nil
	}
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	if namespace = strings.TrimSpace(namespace); namespace == "" {
		namespace = "default"
	}
	surface = strings.TrimSpace(surface)
	segmentID = strings.TrimSpace(segmentID)

	db := txpostgres.FromCtx(ctx, s.Pool)
	const q = `
select v.rules
  from tenant_rules_current c
  join tenant_rule_versions v on v.id = c.rules_version_id
 where c.tenant_id = $1
`
	var raw []byte
	if err := db.QueryRow(ctx, q, tenantID).Scan(&raw); err != nil {
		if txpostgres.IsNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	payloads, err := parseRulesPayload(raw)
	if err != nil {
		return nil, err
	}
	out := make([]rules.Rule, 0, len(payloads))
	for _, payload := range payloads {
		rule, ok := buildRule(payload, orgID, namespace, surface, segmentID, ts)
		if !ok {
			continue
		}
		out = append(out, rule)
	}
	return out, nil
}

func parseRulesPayload(raw []byte) ([]rulePayload, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	switch raw[0] {
	case '[':
		var payloads []rulePayload
		if err := json.Unmarshal(raw, &payloads); err != nil {
			return nil, err
		}
		return payloads, nil
	case '{':
		var envelope map[string]json.RawMessage
		if err := json.Unmarshal(raw, &envelope); err != nil {
			return nil, err
		}
		for _, key := range []string{"rules", "items", "data"} {
			if v, ok := envelope[key]; ok {
				return parseRulesPayload(v)
			}
		}
		var single rulePayload
		if err := json.Unmarshal(raw, &single); err != nil {
			return nil, err
		}
		if len(single) == 0 {
			return nil, nil
		}
		return []rulePayload{single}, nil
	default:
		return nil, nil
	}
}

func buildRule(payload rulePayload, orgID uuid.UUID, namespace, surface, segmentID string, now time.Time) (rules.Rule, bool) {
	if payload == nil {
		return rules.Rule{}, false
	}
	action, ok := normalizeAction(getString(payload, "action", "rule_action", "operation"))
	if !ok {
		return rules.Rule{}, false
	}
	targetType, targetKey, itemIDs := extractTarget(payload)
	if targetType == "" && len(itemIDs) > 0 {
		targetType = rules.RuleTargetItem
	}
	if targetType == "" {
		return rules.Rule{}, false
	}
	if targetType == rules.RuleTargetItem {
		if len(itemIDs) == 0 && targetKey != "" {
			itemIDs = []string{targetKey}
			targetKey = ""
		}
		if len(itemIDs) == 0 {
			return rules.Rule{}, false
		}
	} else if strings.TrimSpace(targetKey) == "" {
		return rules.Rule{}, false
	}

	ruleNamespace := strings.TrimSpace(getString(payload, "namespace", "ns"))
	ruleSurface := strings.TrimSpace(getString(payload, "surface", "surf"))
	ruleSegment := strings.TrimSpace(getString(payload, "segment_id", "segment", "segmentId"))
	if !matchScope(ruleNamespace, ruleSurface, ruleSegment, namespace, surface, segmentID) {
		return rules.Rule{}, false
	}

	enabled := true
	if v, ok := getBool(payload, "enabled", "active", "is_enabled"); ok {
		enabled = v
	}
	if !enabled {
		return rules.Rule{}, false
	}

	var validFrom *time.Time
	if t, ok := getTime(payload, "valid_from", "starts_at", "start_at", "start"); ok {
		validFrom = &t
		if now.Before(t) {
			return rules.Rule{}, false
		}
	}
	var validUntil *time.Time
	if t, ok := getTime(payload, "valid_until", "ends_at", "end_at", "end"); ok {
		validUntil = &t
		if !now.Before(t) {
			return rules.Rule{}, false
		}
	}

	ruleID := parseRuleID(payload, orgID)
	var overrideID *uuid.UUID
	if raw := getString(payload, "manual_override_id", "override_id", "manualOverrideId"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			overrideID = &id
		}
	}

	if ruleNamespace == "" {
		ruleNamespace = namespace
	}
	if ruleSurface == "" {
		ruleSurface = surface
	}

	priority := 0
	if v, ok := getInt(payload, "priority", "order", "rank"); ok {
		priority = v
	}

	var boostValue *float64
	if action == rules.RuleActionBoost {
		if v, ok := getFloat(payload, "boost_value", "boost", "value", "boostValue"); ok {
			boostValue = &v
		}
	}

	var maxPins *int
	if action == rules.RuleActionPin {
		if v, ok := getInt(payload, "max_pins", "maxPins", "max_pinned"); ok {
			maxPins = &v
		}
	}

	createdAt := time.Time{}
	if t, ok := getTime(payload, "created_at", "createdAt"); ok {
		createdAt = t
	}
	updatedAt := createdAt
	if t, ok := getTime(payload, "updated_at", "updatedAt"); ok {
		updatedAt = t
	}

	itemIDs = normalizeStringList(itemIDs)

	return rules.Rule{
		RuleID:           ruleID,
		ManualOverrideID: overrideID,
		OrgID:            orgID,
		Namespace:        ruleNamespace,
		Surface:          ruleSurface,
		Name:             strings.TrimSpace(getString(payload, "name", "title")),
		Description:      strings.TrimSpace(getString(payload, "description", "desc")),
		Action:           action,
		TargetType:       targetType,
		TargetKey:        strings.TrimSpace(targetKey),
		ItemIDs:          itemIDs,
		BoostValue:       boostValue,
		MaxPins:          maxPins,
		SegmentID:        ruleSegment,
		Priority:         priority,
		Enabled:          enabled,
		ValidFrom:        validFrom,
		ValidUntil:       validUntil,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, true
}

func matchScope(ruleNamespace, ruleSurface, ruleSegment, namespace, surface, segmentID string) bool {
	if ruleNamespace != "" && !strings.EqualFold(ruleNamespace, namespace) {
		return false
	}
	if ruleSurface != "" && !strings.EqualFold(ruleSurface, surface) {
		return false
	}
	if ruleSegment != "" {
		if segmentID == "" || !strings.EqualFold(ruleSegment, segmentID) {
			return false
		}
	}
	return true
}

func extractTarget(payload rulePayload) (rules.RuleTarget, string, []string) {
	targetType, _ := normalizeTargetType(getString(payload, "target_type", "targetType"))
	targetKey := getString(payload, "target_key", "targetKey", "key")
	itemIDs := getStringSlice(payload, "item_ids", "item_id", "itemIds", "itemId", "items", "target_items", "targetItems")

	if targetKey == "" {
		if tag := getString(payload, "tag"); tag != "" {
			targetKey = tag
			if targetType == "" {
				targetType = rules.RuleTargetTag
			}
		}
	}
	if targetKey == "" {
		if brand := getString(payload, "brand"); brand != "" {
			targetKey = brand
			if targetType == "" {
				targetType = rules.RuleTargetBrand
			}
		}
	}
	if targetKey == "" {
		if category := getString(payload, "category"); category != "" {
			targetKey = category
			if targetType == "" {
				targetType = rules.RuleTargetCategory
			}
		}
	}

	if target := payload["target"]; target != nil {
		switch v := target.(type) {
		case map[string]any:
			if targetType == "" {
				targetType, _ = normalizeTargetType(getString(v, "type", "target_type", "targetType"))
			}
			if targetKey == "" {
				targetKey = getString(v, "key", "target_key", "tag", "value")
			}
			if len(itemIDs) == 0 {
				itemIDs = getStringSlice(v, "item_ids", "itemIds", "items", "target_items", "targetItems")
			}
		case string:
			if targetType == "" {
				targetType, _ = normalizeTargetType(v)
			}
		}
	}

	if targetType == "" {
		if t, ok := normalizeTargetType(getString(payload, "target")); ok {
			targetType = t
		}
	}

	return targetType, targetKey, itemIDs
}

func normalizeAction(raw string) (rules.RuleAction, bool) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "BLOCK", "BLOCKLIST", "BAN", "EXCLUDE", "SUPPRESS":
		return rules.RuleActionBlock, true
	case "PIN", "PINNED", "FORCE":
		return rules.RuleActionPin, true
	case "BOOST", "PROMOTE", "INCREASE":
		return rules.RuleActionBoost, true
	default:
		return "", false
	}
}

func normalizeTargetType(raw string) (rules.RuleTarget, bool) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "ITEM", "ITEMS", "PRODUCT":
		return rules.RuleTargetItem, true
	case "TAG", "LABEL":
		return rules.RuleTargetTag, true
	case "BRAND":
		return rules.RuleTargetBrand, true
	case "CATEGORY", "CAT":
		return rules.RuleTargetCategory, true
	default:
		return "", false
	}
}

func parseRuleID(payload rulePayload, orgID uuid.UUID) uuid.UUID {
	if raw := getString(payload, "rule_id", "id", "ruleId"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			return id
		}
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		return uuid.New()
	}
	namespace := uuid.NameSpaceOID
	if orgID != uuid.Nil {
		namespace = orgID
	}
	seed := append([]byte(orgID.String()+"|"), blob...)
	return uuid.NewSHA1(namespace, seed)
}

func getString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			if s := stringFromAny(v); s != "" {
				return s
			}
		}
	}
	return ""
}

func getStringSlice(payload map[string]any, keys ...string) []string {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			if items := stringSliceFromAny(v); len(items) > 0 {
				return items
			}
		}
	}
	return nil
}

func getBool(payload map[string]any, keys ...string) (bool, bool) {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			if b, ok := boolFromAny(v); ok {
				return b, true
			}
		}
	}
	return false, false
}

func getInt(payload map[string]any, keys ...string) (int, bool) {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			if i, ok := intFromAny(v); ok {
				return i, true
			}
		}
	}
	return 0, false
}

func getFloat(payload map[string]any, keys ...string) (float64, bool) {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			if f, ok := floatFromAny(v); ok {
				return f, true
			}
		}
	}
	return 0, false
}

func getTime(payload map[string]any, keys ...string) (time.Time, bool) {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			if t, ok := timeFromAny(v); ok {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

func normalizeStringList(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func stringFromAny(v any) string {
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case []byte:
		return strings.TrimSpace(string(val))
	case json.Number:
		return strings.TrimSpace(val.String())
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	default:
		return ""
	}
}

func stringSliceFromAny(v any) []string {
	switch val := v.(type) {
	case []any:
		out := make([]string, 0, len(val))
		for _, item := range val {
			if s := stringFromAny(item); s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return append([]string(nil), val...)
	case string:
		trimmed := strings.TrimSpace(val)
		if trimmed == "" {
			return nil
		}
		if strings.Contains(trimmed, ",") {
			parts := strings.Split(trimmed, ",")
			out := make([]string, 0, len(parts))
			for _, part := range parts {
				if s := strings.TrimSpace(part); s != "" {
					out = append(out, s)
				}
			}
			return out
		}
		return []string{trimmed}
	default:
		return nil
	}
}

func boolFromAny(v any) (bool, bool) {
	switch val := v.(type) {
	case bool:
		return val, true
	case string:
		trimmed := strings.TrimSpace(val)
		if trimmed == "" {
			return false, false
		}
		if parsed, err := strconv.ParseBool(trimmed); err == nil {
			return parsed, true
		}
	case float64:
		return val != 0, true
	case int:
		return val != 0, true
	}
	return false, false
}

func intFromAny(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	case json.Number:
		if i, err := val.Int64(); err == nil {
			return int(i), true
		}
	case string:
		trimmed := strings.TrimSpace(val)
		if trimmed == "" {
			return 0, false
		}
		if i, err := strconv.Atoi(trimmed); err == nil {
			return i, true
		}
	}
	return 0, false
}

func floatFromAny(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		if f, err := val.Float64(); err == nil {
			return f, true
		}
	case string:
		trimmed := strings.TrimSpace(val)
		if trimmed == "" {
			return 0, false
		}
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func timeFromAny(v any) (time.Time, bool) {
	switch val := v.(type) {
	case time.Time:
		if val.IsZero() {
			return time.Time{}, false
		}
		return val, true
	case string:
		trimmed := strings.TrimSpace(val)
		if trimmed == "" {
			return time.Time{}, false
		}
		if t, err := time.Parse(time.RFC3339Nano, trimmed); err == nil {
			return t, true
		}
		if t, err := time.Parse(time.RFC3339, trimmed); err == nil {
			return t, true
		}
		if t, err := time.Parse("2006-01-02", trimmed); err == nil {
			return t, true
		}
	case float64:
		return unixFromFloat(val), true
	case json.Number:
		if i, err := val.Int64(); err == nil {
			return time.Unix(i, 0).UTC(), true
		}
		if f, err := val.Float64(); err == nil {
			return unixFromFloat(f), true
		}
	}
	return time.Time{}, false
}

func unixFromFloat(value float64) time.Time {
	sec := value
	if sec > 1e15 {
		sec = sec / 1e9
	} else if sec > 1e12 {
		sec = sec / 1e3
	}
	whole := int64(sec)
	frac := sec - float64(whole)
	nanos := int64(frac * 1e9)
	return time.Unix(whole, nanos).UTC()
}

var _ rules.Store = (*RulesManagerStore)(nil)
