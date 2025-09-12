package bandit

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

// Store abstracts persistent operations required by the bandit.
type Store interface {
	ListActivePolicies(ctx context.Context, orgID string, ns string) ([]PolicyConfig, error)
	ListPoliciesByIDs(ctx context.Context, orgID, ns string, ids []string) ([]PolicyConfig, error)
	GetStats(ctx context.Context, orgID, ns, surface, bucket string, algo Algorithm) (map[string]Stats, error)
	IncrementStats(ctx context.Context, orgID, ns, surface, bucket string, algo Algorithm, policyID string, reward bool) error
	LogDecision(ctx context.Context, orgID, ns, surface, bucket string, algo Algorithm, policyID string, explore bool, reqID string, meta map[string]any) error
	LogReward(ctx context.Context, orgID, ns, surface, bucket string, algo Algorithm, policyID string, reward bool, reqID string, meta map[string]any) error
}

type Manager struct {
	Store Store
	Algo  Algorithm
	Rand  *rand.Rand
}

func NewManager(s Store, algo Algorithm) *Manager {
	return &Manager{
		Store: s,
		Algo:  algo,
		Rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Decide chooses the best policy for this (surface, context bucket).
// If candidateIDs is empty, all active policies are eligible.
func (m *Manager) Decide(
	ctx context.Context,
	orgID string,
	ns string,
	surface string,
	bucketKey string,
	candidateIDs []string,
	reqID string,
) (Decision, error) {
	var (
		policies []PolicyConfig
		err      error
	)
	if len(candidateIDs) > 0 {
		policies, err = m.Store.ListPoliciesByIDs(ctx, orgID, ns, candidateIDs)
	} else {
		policies, err = m.Store.ListActivePolicies(ctx, orgID, ns)
	}
	if err != nil {
		return Decision{}, err
	}
	if len(policies) == 0 {
		return Decision{}, errors.New("no eligible policies")
	}

	stats, err := m.Store.GetStats(ctx, orgID, ns, surface, bucketKey, m.Algo)
	if err != nil {
		return Decision{}, err
	}

	// Compute pick per algorithm.
	var (
		chosen     PolicyConfig
		explore    bool
		bestScore  = math.Inf(-1)
		empBestID  string
		empBestVal = math.Inf(-1)
	)

	// Empirical best for "explore/exploit" explanation.
	for _, p := range policies {
		s := stats[p.PolicyID]
		mean := 0.0
		if s.Trials > 0 {
			mean = float64(s.Successes) / float64(s.Trials)
		}
		if mean > empBestVal {
			empBestVal = mean
			empBestID = p.PolicyID
		}
	}

	switch m.Algo {
	case AlgorithmThompson:
		for _, p := range policies {
			s := stats[p.PolicyID]
			// Prior defaults to 1,1 if unset; stable for cold start.
			alpha := s.Alpha
			if alpha == 0 {
				alpha = 1
			}
			beta := s.Beta
			if beta == 0 {
				beta = 1
			}
			alpha += float64(s.Successes)
			beta += float64(s.Trials - s.Successes)
			// Sample from Beta(alpha, beta) via Gamma trick.
			x := m.gamma(alpha, 1.0)
			y := m.gamma(beta, 1.0)
			score := x / (x + y)
			if score > bestScore {
				bestScore = score
				chosen = p
			}
		}
		explore = (chosen.PolicyID != empBestID)

	case AlgorithmUCB1:
		// UCB1: mean + sqrt(2 ln N / n_i)
		var N int64
		for _, p := range policies {
			N += max64(1, stats[p.PolicyID].Trials)
		}
		if N <= 0 {
			N = int64(len(policies))
		}
		lnN := math.Log(float64(N))
		for _, p := range policies {
			s := stats[p.PolicyID]
			n := float64(max64(1, s.Trials))
			mean := 0.0
			if s.Trials > 0 {
				mean = float64(s.Successes) / float64(s.Trials)
			}
			bonus := math.Sqrt(2.0 * lnN / n)
			score := mean + bonus
			if score > bestScore {
				bestScore = score
				chosen = p
			}
		}
		// Define "explore" if the optimistic choice differs from empirical best.
		explore = (chosen.PolicyID != empBestID)

	default:
		return Decision{}, errors.New("unknown algorithm")
	}

	_ = m.Store.LogDecision(ctx, orgID, ns, surface, bucketKey, m.Algo,
		chosen.PolicyID, explore, reqID, nil)

	return Decision{
		PolicyID:  chosen.PolicyID,
		Algorithm: m.Algo,
		Surface:   surface,
		BucketKey: bucketKey,
		Explore:   explore,
		Explain: map[string]string{
			"surface":  surface,
			"bucket":   bucketKey,
			"emp_best": empBestID,
		},
	}, nil
}

// Reward updates online stats for the chosen policy.
func (m *Manager) Reward(ctx context.Context, orgID, ns string, in RewardInput, reqID string) error {
	if in.PolicyID == "" || in.Surface == "" || in.BucketKey == "" {
		return errors.New("missing reward fields")
	}
	if in.Algorithm == "" {
		in.Algorithm = m.Algo
	}
	if err := m.Store.IncrementStats(ctx, orgID, ns, in.Surface, in.BucketKey,
		in.Algorithm, in.PolicyID, in.Reward); err != nil {
		return err
	}
	return m.Store.LogReward(ctx, orgID, ns, in.Surface, in.BucketKey,
		in.Algorithm, in.PolicyID, in.Reward, reqID, nil)
}

func (m *Manager) gamma(shape, scale float64) float64 {
	// Marsaglia-Tsang for k>1; simple fallback otherwise.
	k := shape
	if k < 1 {
		k += 1
		x := m.gamma(k, 1.0)
		u := m.Rand.Float64()
		return x * math.Pow(u, 1/shape) * scale
	}
	d := k - 1.0/3.0
	c := 1.0 / math.Sqrt(9.0*d)
	for {
		z := m.norm()
		v := 1 + c*z
		if v <= 0 {
			continue
		}
		v = v * v * v
		u := m.Rand.Float64()
		if u < 1-0.0331*z*z*z*z {
			return d * v * scale
		}
		if math.Log(u) < 0.5*z*z+d*(1-v+math.Log(v)) {
			return d * v * scale
		}
	}
}

func (m *Manager) norm() float64 {
	// Box-Muller
	u1 := m.Rand.Float64()
	u2 := m.Rand.Float64()
	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
