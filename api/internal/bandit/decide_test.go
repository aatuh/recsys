package bandit

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"recsys/internal/types"
)

type stubBanditStore struct {
	policies         []types.PolicyConfig
	stats            map[string]types.Stats
	lastDecisionMeta map[string]any
	lastRewardMeta   map[string]any
}

func (s *stubBanditStore) ListActivePolicies(ctx context.Context, orgID, ns string) ([]types.PolicyConfig, error) {
	return append([]types.PolicyConfig(nil), s.policies...), nil
}

func (s *stubBanditStore) ListPoliciesByIDs(ctx context.Context, orgID, ns string, ids []string) ([]types.PolicyConfig, error) {
	out := make([]types.PolicyConfig, 0, len(ids))
	for _, id := range ids {
		for _, p := range s.policies {
			if p.PolicyID == id {
				out = append(out, p)
				break
			}
		}
	}
	return out, nil
}

func (s *stubBanditStore) GetStats(ctx context.Context, orgID, ns, surface, bucket string, algo types.Algorithm) (map[string]types.Stats, error) {
	return s.stats, nil
}

func (s *stubBanditStore) IncrementStats(ctx context.Context, orgID, ns, surface, bucket string, algo types.Algorithm, policyID string, reward bool) error {
	return nil
}

func (s *stubBanditStore) LogDecision(ctx context.Context, orgID, ns, surface, bucket string, algo types.Algorithm, policyID string, explore bool, reqID string, meta map[string]any) error {
	s.lastDecisionMeta = meta
	return nil
}

func (s *stubBanditStore) LogReward(ctx context.Context, orgID, ns, surface, bucket string, algo types.Algorithm, policyID string, reward bool, reqID string, meta map[string]any) error {
	s.lastRewardMeta = meta
	return nil
}

func TestManagerHoldoutUsesEmpiricalBest(t *testing.T) {
	store := &stubBanditStore{
		policies: []types.PolicyConfig{
			{PolicyID: "control"},
			{PolicyID: "explore"},
		},
		stats: map[string]types.Stats{
			"control": {Trials: 10, Successes: 3},
			"explore": {Trials: 10, Successes: 7},
		},
	}

	mgr := NewManager(store, types.AlgorithmThompson, WithExperiment(ExperimentConfig{
		Enabled:        true,
		HoldoutPercent: 1.0,
		Label:          "rt-7d",
		Surfaces: map[string]struct{}{
			"home": {},
		},
	}))
	mgr.Rand = rand.New(rand.NewSource(42))

	dec, err := mgr.Decide(context.Background(), "org", "ns", "home", "bucket", nil, "req")
	require.NoError(t, err)
	require.Equal(t, "explore", dec.PolicyID)
	require.False(t, dec.Explore)
	require.Equal(t, "rt-7d", dec.Experiment)
	require.Equal(t, "control", dec.Variant)
	require.NotNil(t, store.lastDecisionMeta)
	require.Equal(t, "rt-7d", store.lastDecisionMeta["experiment"])
	require.Equal(t, "control", store.lastDecisionMeta["variant"])
}

func TestManagerRewardForwardsMeta(t *testing.T) {
	store := &stubBanditStore{
		policies: []types.PolicyConfig{{PolicyID: "p1"}},
		stats:    make(map[string]types.Stats),
	}
	mgr := NewManager(store, types.AlgorithmThompson)
	mgr.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	input := RewardInput{
		PolicyID:  "p1",
		Surface:   "home",
		BucketKey: "bucket",
		Reward:    true,
		Meta: map[string]any{
			"experiment": "rt-7d",
			"variant":    "treatment",
		},
	}

	require.NoError(t, mgr.Reward(context.Background(), "org", "ns", input, "req"))
	require.NotNil(t, store.lastRewardMeta)
	require.Equal(t, "treatment", store.lastRewardMeta["variant"])
}
