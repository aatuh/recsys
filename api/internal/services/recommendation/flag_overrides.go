package recommendation

import (
	"context"
	"strings"
	"time"

	"github.com/aatuh/recsys-algo/algorithm"
)

// BlendConfigResolver returns per-namespace override weights.
type BlendConfigResolver interface {
	ResolveBlend(ctx context.Context, namespace string) (*ResolvedBlendConfig, error)
}

// ResolvedBlendConfig captures the evaluated blend parameters and metadata.
type ResolvedBlendConfig struct {
	Namespace string
	Alpha     float64
	Beta      float64
	Gamma     float64
	Source    string
	UpdatedAt time.Time
}

// BlendedConfigFlag implements a simple resolver function.
type BlendedConfigFlag func(ctx context.Context, namespace string) (*ResolvedBlendConfig, error)

func (f BlendedConfigFlag) ResolveBlend(ctx context.Context, namespace string) (*ResolvedBlendConfig, error) {
	return f(ctx, namespace)
}

// StaticBlendResolver provides namespace-specific weights from a fixed map.
type StaticBlendResolver struct {
	overrides map[string]ResolvedBlendConfig
}

// NewStaticBlendResolver constructs a resolver backed by the provided overrides.
// The map key may be a namespace or "*" for a global default.
func NewStaticBlendResolver(entries map[string]ResolvedBlendConfig) *StaticBlendResolver {
	if len(entries) == 0 {
		return nil
	}
	now := time.Now().UTC()
	out := make(map[string]ResolvedBlendConfig, len(entries))
	for ns, cfg := range entries {
		norm := normalizeNamespace(ns)
		cfg.Namespace = norm
		if cfg.Source == "" {
			cfg.Source = "static"
		}
		if cfg.UpdatedAt.IsZero() {
			cfg.UpdatedAt = now
		}
		out[norm] = cfg
	}
	return &StaticBlendResolver{overrides: out}
}

func (s *StaticBlendResolver) ResolveBlend(ctx context.Context, namespace string) (*ResolvedBlendConfig, error) {
	if s == nil || len(s.overrides) == 0 {
		return nil, nil
	}
	norm := normalizeNamespace(namespace)
	if cfg, ok := s.overrides[norm]; ok {
		return &cfg, nil
	}
	if cfg, ok := s.overrides["*"]; ok {
		cfg.Namespace = norm
		return &cfg, nil
	}
	return nil, nil
}

// blendOverrides applies resolved blend values to the algorithm config.
func blendOverrides(cfg *algorithm.Config, resolved *ResolvedBlendConfig) {
	if resolved == nil {
		return
	}
	cfg.BlendAlpha = resolved.Alpha
	cfg.BlendBeta = resolved.Beta
	cfg.BlendGamma = resolved.Gamma
}

func normalizeNamespace(ns string) string {
	ns = strings.TrimSpace(strings.ToLower(ns))
	if ns == "" {
		return "default"
	}
	return ns
}
