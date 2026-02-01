package license

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type fixedClock struct {
	now time.Time
}

func (f fixedClock) Now() time.Time { return f.now }

func TestManagerStatusMissingFile(t *testing.T) {
	mgr := NewManager(Config{FilePath: ""}, nil)
	mgr.clock = fixedClock{now: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	info, err := mgr.Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != StatusUnknown {
		t.Fatalf("expected status unknown, got %s", info.Status)
	}
	if info.Commercial {
		t.Fatalf("expected commercial false for missing license")
	}
}

func TestManagerStatusValid(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	pub, priv := newEd25519Key(t)
	token := signToken(t, priv, Claims{
		Customer:     "Acme",
		Commercial:   boolPtr(true),
		Entitlements: map[string]int{"tenants": 3},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	})

	path := writeLicenseFile(t, token)
	mgr := NewManager(Config{FilePath: path, PublicKey: pub}, nil)
	mgr.clock = fixedClock{now: now}

	info, err := mgr.Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != StatusValid {
		t.Fatalf("expected status valid, got %s", info.Status)
	}
	if !info.Commercial {
		t.Fatalf("expected commercial true")
	}
	if info.Customer != "Acme" {
		t.Fatalf("expected customer Acme, got %s", info.Customer)
	}
	if info.Entitlements["tenants"] != 3 {
		t.Fatalf("expected entitlements tenants=3")
	}
	if info.ExpiresAt == nil || !info.ExpiresAt.Equal(now.Add(24*time.Hour)) {
		t.Fatalf("expected expires_at set")
	}
}

func TestManagerStatusExpired(t *testing.T) {
	now := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	pub, priv := newEd25519Key(t)
	token := signToken(t, priv, Claims{
		Customer: "Acme",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
		},
	})

	path := writeLicenseFile(t, token)
	mgr := NewManager(Config{FilePath: path, PublicKey: pub}, nil)
	mgr.clock = fixedClock{now: now}

	info, err := mgr.Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != StatusExpired {
		t.Fatalf("expected status expired, got %s", info.Status)
	}
	if info.Commercial {
		t.Fatalf("expected commercial false for expired license")
	}
}

func TestManagerStatusInvalidSignature(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	_, priv := newEd25519Key(t)
	pub2, _ := newEd25519Key(t)
	token := signToken(t, priv, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	})

	path := writeLicenseFile(t, token)
	mgr := NewManager(Config{FilePath: path, PublicKey: pub2}, nil)
	mgr.clock = fixedClock{now: now}

	info, err := mgr.Status(context.Background())
	if err == nil {
		t.Fatalf("expected error for invalid signature")
	}
	if info.Status != StatusInvalid {
		t.Fatalf("expected status invalid, got %s", info.Status)
	}
}

func newEd25519Key(t *testing.T) (string, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
	return string(pemBytes), priv
}

func signToken(t *testing.T, priv ed25519.PrivateKey, claims Claims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	str, err := tok.SignedString(priv)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return str
}

func writeLicenseFile(t *testing.T, token string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "license.jwt")
	if err := os.WriteFile(path, []byte(token), 0o600); err != nil {
		t.Fatalf("write license file: %v", err)
	}
	return path
}

func boolPtr(v bool) *bool { return &v }
