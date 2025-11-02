package bandit

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"recsys/internal/types"
	"time"
)

// Manager is the main bandit manager.
type Manager struct {
	Store      types.BanditStore
	Algo       types.Algorithm
	Rand       *rand.Rand
	Experiment ExperimentConfig
}

// Option configures the bandit manager.
type Option func(*Manager)

// WithExperiment enables an exploration experiment.
func WithExperiment(exp ExperimentConfig) Option {
	return func(m *Manager) {
		m.Experiment = exp.Normalized()
	}
}

// NewManager creates a new bandit manager.
func NewManager(s types.BanditStore, algo types.Algorithm, opts ...Option) *Manager {
	mgr := &Manager{
		Store: s,
		Algo:  algo,
		Rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(mgr)
		}
	}
	return mgr
}

// Decide chooses the best policy for this (surface, context bucket).
// If policyIDs is empty, all active policies are eligible.
func (m *Manager) Decide(
	ctx context.Context,
	orgID string,
	ns string,
	surface string,
	bucketKey string,
	policyIDs []string,
	reqID string,
) (Decision, error) {
	var (
		policies []types.PolicyConfig
		err      error
	)
	// Either get the policies by IDs or all active policies.
	if len(policyIDs) > 0 {
		policies, err = m.Store.ListPoliciesByIDs(ctx, orgID, ns, policyIDs)
	} else {
		policies, err = m.Store.ListActivePolicies(ctx, orgID, ns)
	}
	if err != nil {
		return Decision{}, err
	}
	if len(policies) == 0 {
		return Decision{}, errors.New("no eligible policies")
	}

	// Get stats for the chosen algorithm.
	stats, err := m.Store.GetStats(ctx, orgID, ns, surface, bucketKey, m.Algo)
	if err != nil {
		return Decision{}, err
	}

	var (
		chosenPolicy types.PolicyConfig
		explore      bool
		bestScore    = math.Inf(-1)
		empBestID    string
		empBestVal   = math.Inf(-1)
	)

	// Empirical mean best for "explore/exploit" explanation.
	for _, p := range policies {
		s := stats[p.PolicyID]
		mean := 0.0
		if s.Trials > 0 {
			mean = float64(s.Successes) / float64(s.Trials)
		}
		// Keep track of the policy with the highest mean.
		if mean > empBestVal {
			empBestVal = mean
			empBestID = p.PolicyID
		}
	}
	if empBestID == "" && len(policies) > 0 {
		empBestID = policies[0].PolicyID
	}

	holdout := false
	variant := ""
	experimentApplies := m.Experiment.Applies(surface)
	if experimentApplies {
		variant = "treatment"
		if m.Rand.Float64() < m.Experiment.HoldoutPercent {
			holdout = true
			variant = "control"
		}
	}

	switch m.Algo {
	case types.AlgorithmThompson:
		// Thompson sampling: sample from Beta(alpha, beta) via Gamma trick.
		for _, p := range policies {
			s := stats[p.PolicyID]
			// Prior defaults to 1,1 if unset. Stable for cold start.
			// alpha is the number of successes.
			alpha := s.Alpha
			if alpha == 0 {
				alpha = 1
			}
			// beta is the number of failures.
			beta := s.Beta
			if beta == 0 {
				beta = 1
			}
			// Update the prior with the observed successes and failures.
			alpha += float64(s.Successes)
			beta += float64(s.Trials - s.Successes)
			// Thompson: Get CTR ~ Beta(alpha, beta) with these steps:
			// - X ~ Gamma(alpha, 1)
			// - Y ~ Gamma(beta, 1)
			// - X/(X+Y) (Beta(alpha, beta))
			// The ratio is the sampled probability used to rank arms for this
			// decision.
			x := m.gamma(alpha, 1.0)
			y := m.gamma(beta, 1.0)
			score := x / (x + y)
			if score > bestScore {
				bestScore = score
				chosenPolicy = p
			}
		}
		explore = (chosenPolicy.PolicyID != empBestID)

	case types.AlgorithmUCB1:
		// UCB1: meanSuccesses + sqrt(2 * ln(nTrials) / policyTrials)
		var nTrials int64
		// Compute the total number of trials.
		for _, p := range policies {
			nTrials += max64(1, stats[p.PolicyID].Trials)
		}
		if nTrials <= 0 {
			nTrials = int64(len(policies))
		}
		lnN := math.Log(float64(nTrials))
		for _, policy := range policies {
			policyStats := stats[policy.PolicyID]
			policyTrials := float64(max64(1, policyStats.Trials))
			meanSuccesses := 0.0
			if policyStats.Trials > 0 {
				meanSuccesses =
					float64(policyStats.Successes) / float64(policyStats.Trials)
			}
			bonus := math.Sqrt(2.0 * lnN / policyTrials)
			score := meanSuccesses + bonus
			if score > bestScore {
				bestScore = score
				chosenPolicy = policy
			}
		}
		explore = (chosenPolicy.PolicyID != empBestID)

	default:
		return Decision{}, errors.New("unknown algorithm")
	}

	if holdout {
		chosenPolicy = policies[0]
		for _, p := range policies {
			if p.PolicyID == empBestID {
				chosenPolicy = p
				break
			}
		}
		explore = false
	}

	var meta map[string]any
	if experimentApplies {
		meta = map[string]any{
			"experiment":      m.Experiment.Label,
			"variant":         variant,
			"holdout":         holdout,
			"holdout_percent": m.Experiment.HoldoutPercent,
			"surface":         surface,
			"bucket":          bucketKey,
			"empirical_best":  empBestID,
		}
	}

	err = m.Store.LogDecision(
		ctx,
		orgID,
		ns,
		surface,
		bucketKey,
		m.Algo,
		chosenPolicy.PolicyID,
		explore,
		reqID,
		meta,
	)
	if err != nil {
		return Decision{}, err
	}

	return Decision{
		PolicyID:  chosenPolicy.PolicyID,
		Algorithm: m.Algo,
		Surface:   surface,
		BucketKey: bucketKey,
		Explore:   explore,
		Experiment: func() string {
			if experimentApplies {
				return m.Experiment.Label
			}
			return ""
		}(),
		Variant: variant,
		Explain: map[string]string{
			"surface":  surface,
			"bucket":   bucketKey,
			"emp_best": empBestID,
		},
	}, nil
}

// Reward updates online stats for the chosen policy.
func (m *Manager) Reward(
	ctx context.Context, orgID string, ns string, in RewardInput, reqID string,
) error {
	if in.PolicyID == "" || in.Surface == "" || in.BucketKey == "" {
		return errors.New("missing reward fields")
	}
	if in.Algorithm == "" {
		in.Algorithm = m.Algo
	}
	if err := m.Store.IncrementStats(
		ctx,
		orgID,
		ns,
		in.Surface,
		in.BucketKey,
		in.Algorithm,
		in.PolicyID,
		in.Reward,
	); err != nil {
		return err
	}
	return m.Store.LogReward(
		ctx,
		orgID,
		ns,
		in.Surface,
		in.BucketKey,
		in.Algorithm,
		in.PolicyID,
		in.Reward,
		reqID,
		in.Meta,
	)
}

// gamma is a helper function to sample a random variable from a gamma
// distribution.
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

// norm is a helper function to sample from a normal distribution.
func (m *Manager) norm() float64 {
	// Box-Muller
	u1 := m.Rand.Float64()
	u2 := m.Rand.Float64()
	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}

// max64 is a helper function to return the maximum of two int64 values.
func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
