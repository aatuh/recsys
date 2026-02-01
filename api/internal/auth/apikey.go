package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

var (
	ErrAPIKeyNotFound = errors.New("api key not found")
	ErrAPIKeyRevoked  = errors.New("api key revoked")
	ErrAPIKeyExpired  = errors.New("api key expired")
)

// APIKey holds resolved API key metadata.
type APIKey struct {
	ID               string
	TenantID         string
	TenantExternalID string
	Name             string
	Roles            []string
}

// HashAPIKey returns a hex-encoded hash of the provided API key.
// If secret is provided, HMAC-SHA256 is used; otherwise SHA256 is used.
func HashAPIKey(raw, secret string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	secret = strings.TrimSpace(secret)
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write([]byte(raw))
		return hex.EncodeToString(mac.Sum(nil))
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
