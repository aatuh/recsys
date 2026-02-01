package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/aatuh/recsys-suite/api/internal/auth"
	"github.com/aatuh/recsys-suite/api/internal/config"
	"github.com/aatuh/recsys-suite/api/internal/http/problem"

	"github.com/aatuh/api-toolkit/ports"
	"github.com/go-chi/chi/v5"
)

// APIKeyStore resolves API key metadata by hash.
type APIKeyStore interface {
	Lookup(ctx context.Context, hash string) (auth.APIKey, error)
}

// APIKeyMiddleware validates API keys and injects auth context.
type APIKeyMiddleware struct {
	cfg   config.AuthConfig
	store APIKeyStore
	log   ports.Logger
}

// NewAPIKeyMiddleware constructs the middleware.
func NewAPIKeyMiddleware(cfg config.AuthConfig, store APIKeyStore, log ports.Logger) (*APIKeyMiddleware, error) {
	if log == nil {
		log = ports.NopLogger{}
	}
	if !cfg.APIKeys.Enabled {
		return &APIKeyMiddleware{cfg: cfg, store: store, log: log}, nil
	}
	header := strings.TrimSpace(cfg.APIKeys.Header)
	if header == "" {
		return nil, errors.New("api key header is required")
	}
	if store == nil {
		return nil, errors.New("api key store is required")
	}
	cfg.APIKeys.Header = header
	return &APIKeyMiddleware{cfg: cfg, store: store, log: log}, nil
}

// Handler wraps the next handler with API key validation.
func (m *APIKeyMiddleware) Handler(next http.Handler) http.Handler {
	if m == nil {
		return next
	}
	if next == nil {
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isProtectedPath(r) || !m.cfg.APIKeys.Enabled {
			next.ServeHTTP(w, r)
			return
		}
		header := m.cfg.APIKeys.Header
		if header == "" {
			next.ServeHTTP(w, r)
			return
		}
		rawKey := strings.TrimSpace(r.Header.Get(header))
		if rawKey == "" {
			next.ServeHTTP(w, r)
			return
		}
		hash := auth.HashAPIKey(rawKey, m.cfg.APIKeys.HashSecret)
		if hash == "" {
			problem.Write(w, r, http.StatusUnauthorized, "RECSYS_AUTH_INVALID", "invalid api key")
			return
		}
		key, err := m.store.Lookup(r.Context(), hash)
		if err != nil {
			if m.log != nil {
				m.log.Warn("api key auth failed", "err", err.Error())
			}
			problem.Write(w, r, http.StatusUnauthorized, "RECSYS_AUTH_INVALID", "invalid api key")
			return
		}
		info, _ := auth.FromContext(r.Context())
		resolvedTenant, err := m.resolveTenantScope(r, key, info.TenantID)
		if err != nil {
			problem.Write(w, r, http.StatusForbidden, "RECSYS_TENANT_SCOPE_MISMATCH", "tenant scope mismatch")
			return
		}
		info.APIKeyID = key.ID
		info.APIKeyName = key.Name
		if info.UserID == "" {
			info.UserID = "api-key:" + key.ID
		}
		if info.TenantID == "" {
			info.TenantID = resolvedTenant
			info.TenantSource = auth.TenantSourceAPIKey
		}
		info.Roles = mergeRoles(info.Roles, key.Roles)
		ctx := auth.WithInfo(r.Context(), info)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *APIKeyMiddleware) resolveTenantScope(r *http.Request, key auth.APIKey, existing string) (string, error) {
	allowed := map[string]struct{}{}
	if key.TenantID != "" {
		allowed[key.TenantID] = struct{}{}
	}
	if key.TenantExternalID != "" {
		allowed[key.TenantExternalID] = struct{}{}
	}
	if existing != "" && !tenantAllowed(existing, allowed) {
		return "", errors.New("tenant mismatch")
	}

	headerTenant := ""
	if header := strings.TrimSpace(m.cfg.TenantHeader); header != "" {
		val, err := tenantFromHeader(r, header)
		if err != nil {
			return "", err
		}
		headerTenant = val
	}
	urlTenant := strings.TrimSpace(chi.URLParam(r, "tenant_id"))
	if urlTenant != "" && strings.ContainsAny(urlTenant, " \t") {
		return "", errors.New("tenant url param invalid")
	}

	if headerTenant != "" && urlTenant != "" && headerTenant != urlTenant {
		return "", errors.New("tenant mismatch")
	}
	if headerTenant != "" && !tenantAllowed(headerTenant, allowed) {
		return "", errors.New("tenant mismatch")
	}
	if urlTenant != "" && !tenantAllowed(urlTenant, allowed) {
		return "", errors.New("tenant mismatch")
	}
	if existing != "" && headerTenant != "" && existing != headerTenant {
		return "", errors.New("tenant mismatch")
	}
	if existing != "" && urlTenant != "" && existing != urlTenant {
		return "", errors.New("tenant mismatch")
	}

	resolved := urlTenant
	if resolved == "" {
		resolved = headerTenant
	}
	if resolved == "" {
		if existing != "" {
			resolved = existing
		} else if key.TenantExternalID != "" {
			resolved = key.TenantExternalID
		} else {
			resolved = key.TenantID
		}
	}
	return resolved, nil
}

func tenantAllowed(value string, allowed map[string]struct{}) bool {
	if value == "" {
		return false
	}
	_, ok := allowed[value]
	return ok
}

func tenantFromHeader(r *http.Request, header string) (string, error) {
	if r == nil {
		return "", errors.New("request is nil")
	}
	vals := r.Header.Values(header)
	if len(vals) > 1 {
		return "", errors.New("tenant header invalid")
	}
	if len(vals) == 0 {
		return "", nil
	}
	val := strings.TrimSpace(vals[0])
	if val == "" {
		return "", nil
	}
	if strings.Contains(val, ",") || strings.ContainsAny(val, " \t") {
		return "", errors.New("tenant header invalid")
	}
	return val, nil
}
