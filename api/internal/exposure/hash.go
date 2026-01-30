package exposure

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Hasher pseudonymizes identifiers using HMAC-SHA256.
type Hasher struct {
	salt []byte
}

// NewHasher constructs a Hasher with the provided salt.
func NewHasher(salt string) Hasher {
	return Hasher{salt: []byte(salt)}
}

// Hash returns a hex-encoded hash for the value or empty string when blank.
func (h Hasher) Hash(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return ""
	}
	mac := hmac.New(sha256.New, h.salt)
	_, _ = mac.Write([]byte(v))
	return hex.EncodeToString(mac.Sum(nil))
}

// Subject builds a subject with hashed identifiers.
func (h Hasher) Subject(userID, anonymousID, sessionID string) *Subject {
	subject := Subject{
		UserIDHash:      h.Hash(userID),
		AnonymousIDHash: h.Hash(anonymousID),
		SessionIDHash:   h.Hash(sessionID),
	}
	if subject.UserIDHash == "" && subject.AnonymousIDHash == "" && subject.SessionIDHash == "" {
		return nil
	}
	return &subject
}
