package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ParseBearerToken parses an Authorization header into a bearer token.
func ParseBearerToken(header string) (string, bool, error) {
	if header == "" {
		return "", false, nil
	}
	if strings.Contains(header, ",") {
		return "", true, errors.New("authorization header contains multiple values")
	}
	if header != strings.TrimSpace(header) {
		return "", true, errors.New("authorization header has leading/trailing whitespace")
	}
	const prefix = "bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", true, errors.New("authorization scheme is not bearer")
	}
	token := header[len(prefix):]
	if token == "" {
		return "", true, errors.New("bearer token is empty")
	}
	if strings.ContainsAny(token, " \t") {
		return "", true, errors.New("bearer token contains whitespace")
	}
	return token, true, nil
}

// DecodeJWTClaims extracts JWT claims without signature verification.
// Only call after the token has been validated by the auth middleware.
func DecodeJWTClaims(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, errors.New("token has insufficient segments")
	}
	payload := parts[1]
	data, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("decode claims: %w", err)
	}
	var claims map[string]any
	if err := json.Unmarshal(data, &claims); err != nil {
		return nil, fmt.Errorf("parse claims: %w", err)
	}
	return claims, nil
}

// ExtractTenant pulls the tenant id from the provided claims.
func ExtractTenant(claims map[string]any, keys []string) string {
	for _, key := range keys {
		if key == "" {
			continue
		}
		val := stringClaim(claims, key)
		if strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

// ExtractRoles pulls roles/scopes from the provided claims.
func ExtractRoles(claims map[string]any, keys []string) []string {
	if len(keys) == 0 {
		return nil
	}
	out := []string{}
	seen := map[string]struct{}{}
	for _, key := range keys {
		if key == "" {
			continue
		}
		raw, ok := claimLookup(claims, key)
		if !ok || raw == nil {
			continue
		}
		switch v := raw.(type) {
		case string:
			addRoles(&out, seen, splitRoles(v)...)
		case []string:
			addRoles(&out, seen, v...)
		case []any:
			for _, item := range v {
				switch vv := item.(type) {
				case string:
					addRoles(&out, seen, vv)
				default:
					if str := fmt.Sprint(vv); str != "" {
						addRoles(&out, seen, str)
					}
				}
			}
		default:
			if str := fmt.Sprint(v); str != "" {
				addRoles(&out, seen, splitRoles(str)...)
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func stringClaim(claims map[string]any, key string) string {
	if claims == nil {
		return ""
	}
	val, ok := claimLookup(claims, key)
	if !ok || val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func claimLookup(claims map[string]any, key string) (any, bool) {
	if claims == nil {
		return nil, false
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, false
	}
	if !strings.Contains(key, ".") {
		val, ok := claims[key]
		return val, ok
	}
	parts := strings.Split(key, ".")
	var current any = claims
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, false
		}
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		val, ok := m[part]
		if !ok {
			return nil, false
		}
		current = val
	}
	return current, true
}

func splitRoles(input string) []string {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return nil
	}
	parts := strings.Fields(raw)
	if len(parts) > 1 {
		return parts
	}
	parts = strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if val := strings.TrimSpace(part); val != "" {
			out = append(out, val)
		}
	}
	return out
}

func addRoles(dst *[]string, seen map[string]struct{}, roles ...string) {
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		key := strings.ToLower(role)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		*dst = append(*dst, role)
	}
}
