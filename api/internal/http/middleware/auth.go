package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"recsys/internal/http/common"
)

type apiKeyContextKey struct{}

// APIKeyAccess defines the authorization scope associated with an API key.
type APIKeyAccess struct {
	AllowAll bool
	OrgIDs   map[uuid.UUID]struct{}
}

// APIKeyAuthorizer enforces API key authentication and tenant scoping.
type APIKeyAuthorizer struct {
	keys   map[string]APIKeyAccess
	logger *zap.Logger
}

// NewAPIKeyAuthorizer constructs an authorizer backed by the provided key map.
func NewAPIKeyAuthorizer(keys map[string]APIKeyAccess, logger *zap.Logger) *APIKeyAuthorizer {
	sanitized := make(map[string]APIKeyAccess, len(keys))
	for k, v := range keys {
		if k == "" {
			continue
		}
		entry := APIKeyAccess{
			AllowAll: v.AllowAll,
		}
		if !v.AllowAll && len(v.OrgIDs) > 0 {
			entry.OrgIDs = make(map[uuid.UUID]struct{}, len(v.OrgIDs))
			for id := range v.OrgIDs {
				entry.OrgIDs[id] = struct{}{}
			}
		}
		sanitized[k] = entry
	}
	return &APIKeyAuthorizer{keys: sanitized, logger: logger}
}

// Middleware returns the HTTP middleware enforcing API key presence and scope.
func (a *APIKeyAuthorizer) Middleware(next http.Handler) http.Handler {
	if a == nil || len(a.keys) == 0 {
		return next
	}
	logger := a.logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			common.Unauthorized(w, r)
			return
		}

		access, ok := a.keys[apiKey]
		if !ok {
			logger.Warn("invalid api key", zap.String("path", r.URL.Path))
			common.Unauthorized(w, r)
			return
		}

		orgID, ok := OrgIDFromContext(r.Context())
		if !ok {
			err := errors.New("missing tenant context; RequireOrgID middleware must run first")
			common.HttpErrorWithLogger(w, r, err, http.StatusInternalServerError, logger)
			return
		}

		if !access.AllowAll {
			if _, allowed := access.OrgIDs[orgID]; !allowed {
				logger.Warn("api key not allowed for tenant", zap.String("path", r.URL.Path), zap.String("org_id", orgID.String()))
				common.Forbidden(w, r)
				return
			}
		}

		ctx := context.WithValue(r.Context(), apiKeyContextKey{}, apiKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// APIKeyFromContext retrieves the authenticated API key from request context.
func APIKeyFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	if v, ok := ctx.Value(apiKeyContextKey{}).(string); ok {
		return v, true
	}
	return "", false
}

// ContextWithAPIKey decorates the provided context with an API key.
func ContextWithAPIKey(ctx context.Context, key string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, apiKeyContextKey{}, key)
}
