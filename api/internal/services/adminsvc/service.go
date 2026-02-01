package adminsvc

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aatuh/recsys-suite/api/internal/admin"

	"github.com/google/uuid"
)

// Store abstracts admin persistence operations.
type Store interface {
	GetTenantConfig(ctx context.Context, tenantID string) (TenantConfig, error)
	GetTenantRules(ctx context.Context, tenantID string) (TenantRules, error)
	UpdateTenantConfig(ctx context.Context, tenantID string, raw []byte, ifMatch string, actor Actor, meta RequestMeta) (TenantConfig, error)
	UpdateTenantRules(ctx context.Context, tenantID string, raw []byte, ifMatch string, actor Actor, meta RequestMeta) (TenantRules, error)
	ResolveTenantID(ctx context.Context, tenantID string) (uuid.UUID, error)
	InsertCacheInvalidation(ctx context.Context, event CacheInvalidationEvent) error
	ListAuditLog(ctx context.Context, tenantID string, query AuditQuery) (AuditLog, error)
	InsertAuditEvent(ctx context.Context, event AuditEvent) error
}

// ConfigCache supports cache invalidation for tenant configs.
type ConfigCache interface {
	Invalidate(tenantID, surface string) int
}

// RulesCache supports cache invalidation for tenant rules.
type RulesCache interface {
	Invalidate(tenantID, surface string) int
}

// ArtifactCache supports cache invalidation for artifact-backed stores.
type ArtifactCache interface {
	Invalidate(tenantID, surface string) int
}

// RulesManager invalidates rule manager caches.
type RulesManager interface {
	Invalidate(orgID uuid.UUID, namespace, surface string)
}

// Service provides admin/config/rules operations.
type Service struct {
	store            Store
	configCache      ConfigCache
	rulesCache       RulesCache
	artifactCache    ArtifactCache
	rulesManager     RulesManager
	defaultNamespace string
}

const (
	DefaultAuditLimit = 100
	MaxAuditLimit     = 200
)

// ServiceOption customizes the admin service.
type ServiceOption func(*Service)

// WithConfigCache wires a config cache for invalidations.
func WithConfigCache(cache ConfigCache) ServiceOption {
	return func(s *Service) {
		s.configCache = cache
	}
}

// WithRulesCache wires a rules cache for invalidations.
func WithRulesCache(cache RulesCache) ServiceOption {
	return func(s *Service) {
		s.rulesCache = cache
	}
}

// WithArtifactCache wires an artifact cache for invalidations.
func WithArtifactCache(cache ArtifactCache) ServiceOption {
	return func(s *Service) {
		s.artifactCache = cache
	}
}

// WithRulesManager wires a rules manager for invalidations.
func WithRulesManager(manager RulesManager, defaultNamespace string) ServiceOption {
	return func(s *Service) {
		s.rulesManager = manager
		s.defaultNamespace = strings.TrimSpace(defaultNamespace)
	}
}

// New constructs an admin service.
func New(store Store, opts ...ServiceOption) *Service {
	svc := &Service{store: store}
	for _, opt := range opts {
		if opt != nil {
			opt(svc)
		}
	}
	return svc
}

// GetTenantConfig fetches the current tenant config.
func (s *Service) GetTenantConfig(ctx context.Context, tenantID string) (TenantConfig, error) {
	if s == nil || s.store == nil {
		return TenantConfig{}, admin.ErrConfigNotFound
	}
	return s.store.GetTenantConfig(ctx, tenantID)
}

// UpdateTenantConfig updates tenant config with optimistic concurrency.
func (s *Service) UpdateTenantConfig(ctx context.Context, tenantID string, raw []byte, ifMatch string, actor Actor, meta RequestMeta) (TenantConfig, error) {
	if s == nil || s.store == nil {
		return TenantConfig{}, admin.ErrConfigNotFound
	}
	cfg, err := s.store.UpdateTenantConfig(ctx, tenantID, raw, ifMatch, actor, meta)
	if err != nil {
		return TenantConfig{}, err
	}
	if s.configCache != nil {
		s.configCache.Invalidate(tenantID, "")
	}
	return cfg, nil
}

// GetTenantRules fetches the current tenant rules.
func (s *Service) GetTenantRules(ctx context.Context, tenantID string) (TenantRules, error) {
	if s == nil || s.store == nil {
		return TenantRules{}, admin.ErrRulesNotFound
	}
	return s.store.GetTenantRules(ctx, tenantID)
}

// UpdateTenantRules updates tenant rules with optimistic concurrency.
func (s *Service) UpdateTenantRules(ctx context.Context, tenantID string, raw []byte, ifMatch string, actor Actor, meta RequestMeta) (TenantRules, error) {
	if s == nil || s.store == nil {
		return TenantRules{}, admin.ErrRulesNotFound
	}
	rules, err := s.store.UpdateTenantRules(ctx, tenantID, raw, ifMatch, actor, meta)
	if err != nil {
		return TenantRules{}, err
	}
	if s.rulesCache != nil {
		s.rulesCache.Invalidate(tenantID, "")
	}
	if s.rulesManager != nil {
		if orgID, err := s.store.ResolveTenantID(ctx, tenantID); err == nil && orgID != uuid.Nil {
			s.rulesManager.Invalidate(orgID, "", "")
		}
	}
	return rules, nil
}

// InvalidateCache invalidates tenant caches and persists the request.
func (s *Service) InvalidateCache(ctx context.Context, tenantID string, req CacheInvalidateRequest, actor Actor, meta RequestMeta) (CacheInvalidateResult, error) {
	if s == nil || s.store == nil {
		return CacheInvalidateResult{}, admin.ErrTenantNotFound
	}
	tenantID = strings.TrimSpace(tenantID)
	orgID, err := s.store.ResolveTenantID(ctx, tenantID)
	if err != nil {
		return CacheInvalidateResult{}, err
	}
	surface := strings.TrimSpace(req.Surface)
	targets := normalizeTargets(req.Targets)

	invalidated := map[string]int{}
	for _, target := range targets {
		switch target {
		case "config":
			if s.configCache != nil {
				invalidated[target] = s.configCache.Invalidate(tenantID, surface)
			}
		case "rules":
			if s.rulesCache != nil {
				invalidated[target] = s.rulesCache.Invalidate(tenantID, surface)
			}
			if s.rulesManager != nil && surface != "" {
				ns := strings.TrimSpace(surface)
				if ns == "" {
					ns = s.defaultNamespace
				}
				s.rulesManager.Invalidate(orgID, ns, surface)
			}
		case "popularity":
			if s.artifactCache != nil {
				invalidated[target] = s.artifactCache.Invalidate(tenantID, surface)
			}
		}
	}

	status := "applied"
	var requestUUID *uuid.UUID
	if meta.RequestID != "" {
		if id, err := uuid.Parse(meta.RequestID); err == nil {
			requestUUID = &id
		}
	}
	event := CacheInvalidationEvent{
		TenantID:    orgID,
		RequestID:   requestUUID,
		ActorID:     actor.ID,
		Targets:     targets,
		Surface:     surface,
		Status:      status,
		ErrorDetail: "",
	}
	if err := s.store.InsertCacheInvalidation(ctx, event); err != nil {
		return CacheInvalidateResult{}, err
	}
	if err := s.store.InsertAuditEvent(ctx, AuditEvent{
		TenantID:   orgID,
		Actor:      actor,
		Meta:       meta,
		Action:     "cache.invalidate",
		EntityType: "cache_invalidation",
		EntityID:   "",
		After:      mustJSON(map[string]any{"targets": targets, "surface": surface, "status": status}),
	}); err != nil {
		return CacheInvalidateResult{}, err
	}
	return CacheInvalidateResult{
		TenantID:    tenantID,
		Targets:     targets,
		Surface:     surface,
		Status:      status,
		Invalidated: invalidated,
	}, nil
}

// ListAuditLog returns audit events for a tenant.
func (s *Service) ListAuditLog(ctx context.Context, tenantID string, query AuditQuery) (AuditLog, error) {
	if s == nil || s.store == nil {
		return AuditLog{}, admin.ErrTenantNotFound
	}
	if query.Limit <= 0 {
		query.Limit = DefaultAuditLimit
	}
	if query.Limit > MaxAuditLimit {
		query.Limit = MaxAuditLimit
	}
	return s.store.ListAuditLog(ctx, tenantID, query)
}

func normalizeTargets(targets []string) []string {
	if len(targets) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(targets))
	out := make([]string, 0, len(targets))
	for _, raw := range targets {
		key := strings.ToLower(strings.TrimSpace(raw))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

func mustJSON(payload any) []byte {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return raw
}
