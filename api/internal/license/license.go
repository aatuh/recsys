package license

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aatuh/api-toolkit/v2/ports"
	"github.com/golang-jwt/jwt/v5"
)

// Status describes the license validity state.
type Status string

const (
	StatusUnknown Status = "unknown"
	StatusValid   Status = "valid"
	StatusInvalid Status = "invalid"
	StatusExpired Status = "expired"
)

// Info represents the license status payload returned by the service.
type Info struct {
	Status       Status         `json:"status"`
	Commercial   bool           `json:"commercial"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	Customer     string         `json:"customer,omitempty"`
	Entitlements map[string]int `json:"entitlements,omitempty"`
}

// Claims defines JWT fields for license verification.
type Claims struct {
	Customer     string         `json:"customer,omitempty"`
	Commercial   *bool          `json:"commercial,omitempty"`
	Entitlements map[string]int `json:"entitlements,omitempty"`
	jwt.RegisteredClaims
}

// Config controls license verification.
type Config struct {
	FilePath      string
	PublicKey     string
	PublicKeyFile string
	CacheTTL      time.Duration
}

// Clock enables deterministic tests.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// Manager loads and verifies license payloads.
type Manager struct {
	cfg   Config
	log   ports.Logger
	clock Clock

	mu         sync.Mutex
	cached     Info
	cacheUntil time.Time
	pubKey     any
	pubKeyErr  error
}

// NewManager constructs a license manager with optional caching.
func NewManager(cfg Config, log ports.Logger) *Manager {
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = time.Minute
	}
	return &Manager{cfg: cfg, log: log, clock: realClock{}}
}

// Status returns the current license status.
func (m *Manager) Status(ctx context.Context) (Info, error) {
	_ = ctx
	if m == nil {
		return Info{Status: StatusUnknown}, nil
	}
	now := m.clock.Now()
	if cached, ok := m.fromCache(now); ok {
		return cached, nil
	}
	info, err := m.evaluate(now)
	m.storeCache(now, info)
	return info, err
}

func (m *Manager) fromCache(now time.Time) (Info, bool) {
	if m == nil || m.cfg.CacheTTL <= 0 {
		return Info{}, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.cacheUntil.IsZero() && now.Before(m.cacheUntil) {
		return m.cached, true
	}
	return Info{}, false
}

func (m *Manager) storeCache(now time.Time, info Info) {
	if m == nil || m.cfg.CacheTTL <= 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cached = info
	m.cacheUntil = now.Add(m.cfg.CacheTTL)
}

func (m *Manager) evaluate(now time.Time) (Info, error) {
	path := strings.TrimSpace(m.cfg.FilePath)
	if path == "" {
		return Info{Status: StatusUnknown}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Info{Status: StatusUnknown}, nil
		}
		return Info{Status: StatusUnknown}, err
	}
	raw := strings.TrimSpace(string(data))
	if raw == "" {
		return Info{Status: StatusUnknown}, nil
	}

	tokenStr, err := extractToken(raw)
	if err != nil {
		return Info{Status: StatusInvalid, Commercial: false}, err
	}

	return m.parseToken(tokenStr, now)
}

func extractToken(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("empty license payload")
	}
	if strings.HasPrefix(raw, "{") {
		var envelope struct {
			Token   string `json:"token"`
			JWT     string `json:"jwt"`
			License string `json:"license"`
		}
		if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
			return "", fmt.Errorf("decode license envelope: %w", err)
		}
		token := strings.TrimSpace(envelope.Token)
		if token == "" {
			token = strings.TrimSpace(envelope.JWT)
		}
		if token == "" {
			token = strings.TrimSpace(envelope.License)
		}
		if token == "" {
			return "", errors.New("license envelope missing token")
		}
		return token, nil
	}
	return raw, nil
}

func (m *Manager) parseToken(tokenStr string, now time.Time) (Info, error) {
	key, err := m.publicKey()
	if err != nil {
		return Info{Status: StatusInvalid}, err
	}
	claims := &Claims{}
	options := []jwt.ParserOption{jwt.WithTimeFunc(func() time.Time { return now })}
	if methods := validMethodsForKey(key); len(methods) > 0 {
		options = append(options, jwt.WithValidMethods(methods))
	}
	parser := jwt.NewParser(options...)
	token, parseErr := parser.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (any, error) {
		return key, nil
	})

	status := StatusInvalid
	if parseErr == nil && token != nil && token.Valid {
		status = StatusValid
	} else if errors.Is(parseErr, jwt.ErrTokenExpired) {
		status = StatusExpired
	}

	info := Info{Status: status, Commercial: false}
	if claims.ExpiresAt != nil {
		exp := claims.ExpiresAt.Time
		info.ExpiresAt = &exp
	}
	if claims.Customer != "" {
		info.Customer = claims.Customer
	}
	if len(claims.Entitlements) > 0 {
		info.Entitlements = cloneEntitlements(claims.Entitlements)
	}
	if status == StatusValid {
		if claims.Commercial != nil {
			info.Commercial = *claims.Commercial
		} else {
			info.Commercial = true
		}
	}

	if parseErr != nil && !errors.Is(parseErr, jwt.ErrTokenExpired) && m.log != nil {
		m.log.Warn("license token parse failed", "err", parseErr.Error())
	}
	if errors.Is(parseErr, jwt.ErrTokenExpired) {
		return info, nil
	}
	return info, parseErr
}

func (m *Manager) publicKey() (any, error) {
	m.mu.Lock()
	cached := m.pubKey
	cachedErr := m.pubKeyErr
	m.mu.Unlock()
	if cached != nil || cachedErr != nil {
		return cached, cachedErr
	}
	raw := strings.TrimSpace(m.cfg.PublicKey)
	if raw == "" && strings.TrimSpace(m.cfg.PublicKeyFile) != "" {
		data, err := os.ReadFile(m.cfg.PublicKeyFile)
		if err != nil {
			m.setKey(nil, fmt.Errorf("read public key file: %w", err))
			return nil, m.pubKeyErr
		}
		raw = strings.TrimSpace(string(data))
	}
	if raw == "" {
		m.setKey(nil, errors.New("license public key not configured"))
		return nil, m.pubKeyErr
	}
	key, err := parsePublicKey([]byte(raw))
	m.setKey(key, err)
	return key, err
}

func (m *Manager) setKey(key any, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pubKey = key
	m.pubKeyErr = err
}

func parsePublicKey(data []byte) (any, error) {
	if key, err := jwt.ParseEdPublicKeyFromPEM(data); err == nil {
		return key, nil
	}
	if key, err := jwt.ParseRSAPublicKeyFromPEM(data); err == nil {
		return key, nil
	}
	if key, err := jwt.ParseECPublicKeyFromPEM(data); err == nil {
		return key, nil
	}
	return nil, errors.New("unsupported public key format")
}

func validMethodsForKey(key any) []string {
	switch key.(type) {
	case ed25519.PublicKey:
		return []string{jwt.SigningMethodEdDSA.Alg()}
	case *rsa.PublicKey:
		return []string{jwt.SigningMethodRS256.Alg(), jwt.SigningMethodRS384.Alg(), jwt.SigningMethodRS512.Alg()}
	case *ecdsa.PublicKey:
		return []string{jwt.SigningMethodES256.Alg(), jwt.SigningMethodES384.Alg(), jwt.SigningMethodES512.Alg()}
	default:
		return nil
	}
}

func cloneEntitlements(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
