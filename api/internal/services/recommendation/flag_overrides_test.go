package recommendation

import (
	"context"
	"testing"

	spectypes "recsys/specs/types"

	"github.com/aatuh/recsys-algo/algorithm"

	"github.com/stretchr/testify/require"
)

func TestStaticBlendResolver(t *testing.T) {
	entries := map[string]ResolvedBlendConfig{
		"default": {Alpha: 0.1, Beta: 0.2, Gamma: 0.7},
		"*":       {Alpha: 0.3, Beta: 0.3, Gamma: 0.4},
	}
	resolver := NewStaticBlendResolver(entries)
	require.NotNil(t, resolver)

	res, err := resolver.ResolveBlend(context.Background(), "default")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.InDelta(t, 0.1, res.Alpha, 1e-6)
	require.Equal(t, "default", res.Namespace)

	resFallback, err := resolver.ResolveBlend(context.Background(), "unknown_ns")
	require.NoError(t, err)
	require.NotNil(t, resFallback)
	require.InDelta(t, 0.3, resFallback.Alpha, 1e-6)
	require.Equal(t, "unknown_ns", resFallback.Namespace)
}

func TestBlendOverridesRespectRequestOverride(t *testing.T) {
	cfg := algorithm.Config{BlendAlpha: 0.5, BlendBeta: 0.3, BlendGamma: 0.2}
	resolved := &ResolvedBlendConfig{Alpha: 0.1, Beta: 0.2, Gamma: 0.7}
	blendOverrides(&cfg, resolved)
	require.InDelta(t, 0.1, cfg.BlendAlpha, 1e-6)

	alpha := 0.9
	overrides := &spectypes.Overrides{BlendAlpha: &alpha}
	applyOverrides(&cfg, overrides)
	require.InDelta(t, 0.9, cfg.BlendAlpha, 1e-6)
}

func TestServiceResolveBlendNilResolver(t *testing.T) {
	svc := New(nil, nil)
	require.Nil(t, svc.resolveBlend(context.Background(), "default"))

	entries := map[string]ResolvedBlendConfig{"default": {Alpha: 0.4, Beta: 0.4, Gamma: 0.2}}
	resolver := NewStaticBlendResolver(entries)
	svc.WithBlendResolver(resolver)
	res := svc.resolveBlend(context.Background(), "default")
	require.NotNil(t, res)
	require.InDelta(t, 0.4, res.Alpha, 1e-6)
}
