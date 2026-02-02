package recsysvc

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/aatuh/api-toolkit/v2/authorization"
	"github.com/aatuh/recsys-suite/api/internal/artifacts"
)

// Engine defines the algorithm interface for recommendations.
type Engine interface {
	Recommend(ctx context.Context, req RecommendRequest) ([]Item, []Warning, error)
	Similar(ctx context.Context, req SimilarRequest) ([]Item, []Warning, error)
}

// Service orchestrates recommendation requests.
type Service struct {
	engine      Engine
	queue       *BoundedQueue
	configStore ConfigStore
	rulesStore  RulesStore
}

// AlgoVersioner allows engines to report the active algorithm version per request.
type AlgoVersioner interface {
	VersionForRecommend(ctx context.Context, req RecommendRequest) string
	VersionForSimilar(ctx context.Context, req SimilarRequest) string
}

// ServiceOption configures the service.
type ServiceOption func(*Service)

// WithBackpressure sets the backpressure queue.
func WithBackpressure(queue *BoundedQueue) ServiceOption {
	return func(s *Service) {
		s.queue = queue
	}
}

// WithConfigStore sets the tenant config store.
func WithConfigStore(store ConfigStore) ServiceOption {
	return func(s *Service) {
		s.configStore = store
	}
}

// WithRulesStore sets the tenant rules store.
func WithRulesStore(store RulesStore) ServiceOption {
	return func(s *Service) {
		s.rulesStore = store
	}
}

// New constructs a new Service.
func New(engine Engine) *Service {
	return NewWithOptions(engine)
}

// NewWithOptions constructs a new Service with optional dependencies.
func NewWithOptions(engine Engine, opts ...ServiceOption) *Service {
	svc := &Service{engine: engine}
	for _, opt := range opts {
		if opt != nil {
			opt(svc)
		}
	}
	return svc
}

// Recommend returns ranked recommendations.
func (s *Service) Recommend(ctx context.Context, req RecommendRequest) ([]Item, []Warning, ResponseMeta, error) {
	meta := ResponseMeta{AlgoVersion: s.algoVersion()}
	if s == nil || s.engine == nil {
		return nil, nil, meta, nil
	}
	if err := s.acquire(ctx); err != nil {
		return nil, nil, meta, err
	}
	defer s.release()
	if cfg, err := s.loadTenantConfig(ctx, req.Surface); err != nil {
		return nil, nil, meta, err
	} else if cfg.Version != "" {
		meta.ConfigVersion = cfg.Version
		if req.Weights == nil && cfg.Weights != nil {
			req.Weights = cfg.Weights
		}
		if req.Algorithm == "" && cfg.Algo != "" {
			req.Algorithm = cfg.Algo
		}
	}
	if v, ok := s.engine.(AlgoVersioner); ok {
		meta.AlgoVersion = v.VersionForRecommend(ctx, req)
	}
	if rules, err := s.loadTenantRules(ctx, req.Surface); err != nil {
		return nil, nil, meta, err
	} else if rules.Version != "" {
		meta.RulesVersion = rules.Version
	}
	items, warnings, err := s.engine.Recommend(ctx, req)
	if err != nil {
		if errors.Is(err, artifacts.ErrArtifactIncompatible) || errors.Is(err, artifacts.ErrManifestIncompatible) {
			return nil, warnings, meta, ErrArtifactIncompatible
		}
		return nil, warnings, meta, err
	}
	applyDeterministicOrdering(items)
	return items, warnings, meta, nil
}

// Similar returns similar items for a given item.
func (s *Service) Similar(ctx context.Context, req SimilarRequest) ([]Item, []Warning, ResponseMeta, error) {
	meta := ResponseMeta{AlgoVersion: s.algoVersion()}
	if s == nil || s.engine == nil {
		return nil, nil, meta, nil
	}
	if err := s.acquire(ctx); err != nil {
		return nil, nil, meta, err
	}
	defer s.release()
	if cfg, err := s.loadTenantConfig(ctx, req.Surface); err != nil {
		return nil, nil, meta, err
	} else if cfg.Version != "" {
		meta.ConfigVersion = cfg.Version
		if req.Algorithm == "" && cfg.Algo != "" {
			req.Algorithm = cfg.Algo
		}
	}
	if v, ok := s.engine.(AlgoVersioner); ok {
		meta.AlgoVersion = v.VersionForSimilar(ctx, req)
	}
	if rules, err := s.loadTenantRules(ctx, req.Surface); err != nil {
		return nil, nil, meta, err
	} else if rules.Version != "" {
		meta.RulesVersion = rules.Version
	}
	items, warnings, err := s.engine.Similar(ctx, req)
	if err != nil {
		if errors.Is(err, artifacts.ErrArtifactIncompatible) || errors.Is(err, artifacts.ErrManifestIncompatible) {
			return nil, warnings, meta, ErrArtifactIncompatible
		}
		return nil, warnings, meta, err
	}
	applyDeterministicOrdering(items)
	return items, warnings, meta, nil
}

func (s *Service) acquire(ctx context.Context) error {
	if s == nil || s.queue == nil || !s.queue.Enabled() {
		return nil
	}
	return s.queue.Acquire(ctx)
}

func (s *Service) release() {
	if s == nil || s.queue == nil || !s.queue.Enabled() {
		return
	}
	s.queue.Release()
}

func (s *Service) loadTenantConfig(ctx context.Context, surface string) (TenantConfig, error) {
	if s == nil || s.configStore == nil {
		return TenantConfig{}, nil
	}
	tenantID, ok := authorization.TenantIDFromContext(ctx)
	if !ok || tenantID == "" {
		return TenantConfig{}, nil
	}
	cfg, err := s.configStore.GetConfig(ctx, tenantID, surface)
	if err != nil && err != ErrConfigNotFound {
		return TenantConfig{}, err
	}
	if err == ErrConfigNotFound {
		return TenantConfig{}, nil
	}
	return cfg, nil
}

func (s *Service) loadTenantRules(ctx context.Context, surface string) (TenantRules, error) {
	if s == nil || s.rulesStore == nil {
		return TenantRules{}, nil
	}
	tenantID, ok := authorization.TenantIDFromContext(ctx)
	if !ok || tenantID == "" {
		return TenantRules{}, nil
	}
	rules, err := s.rulesStore.GetRules(ctx, tenantID, surface)
	if err != nil && err != ErrRulesNotFound {
		return TenantRules{}, err
	}
	if err == ErrRulesNotFound {
		return TenantRules{}, nil
	}
	return rules, nil
}

// NewNoopEngine returns an engine that always returns empty results.
func NewNoopEngine() Engine {
	return noopEngine{}
}

type noopEngine struct{}

func (noopEngine) Recommend(ctx context.Context, req RecommendRequest) ([]Item, []Warning, error) {
	return nil, nil, nil
}

func (noopEngine) Similar(ctx context.Context, req SimilarRequest) ([]Item, []Warning, error) {
	return nil, nil, nil
}

func (noopEngine) Version() string {
	return defaultAlgoVersion
}

const defaultAlgoVersion = "recsys-algo@stub"

func (s *Service) algoVersion() string {
	if s == nil || s.engine == nil {
		return defaultAlgoVersion
	}
	if v, ok := s.engine.(interface{ Version() string }); ok {
		ver := v.Version()
		if strings.TrimSpace(ver) != "" {
			return ver
		}
	}
	return defaultAlgoVersion
}

func applyDeterministicOrdering(items []Item) {
	if len(items) == 0 {
		return
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return items[i].ItemID < items[j].ItemID
		}
		return items[i].Score > items[j].Score
	})
	for i := range items {
		items[i].Rank = i + 1
	}
}
