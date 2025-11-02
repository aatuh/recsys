package blend

import (
	"context"
	"errors"
	"fmt"
	"time"

	"recsys/internal/algorithm"
	"recsys/internal/store"

	"github.com/google/uuid"
)

// CandidateConfig describes a blend configuration to evaluate.
type CandidateConfig struct {
	Name        string  `yaml:"name" json:"name"`
	Description string  `yaml:"description,omitempty" json:"description,omitempty"`
	Alpha       float64 `yaml:"alpha" json:"alpha"`
	Beta        float64 `yaml:"beta" json:"beta"`
	Gamma       float64 `yaml:"gamma" json:"gamma"`
}

// Validate ensures the candidate configuration is usable.
func (c CandidateConfig) Validate() error {
	if c.Name == "" {
		return errors.New("candidate name is required")
	}
	if c.Alpha < 0 || c.Beta < 0 || c.Gamma < 0 {
		return fmt.Errorf("candidate %q has negative weights", c.Name)
	}
	if c.Alpha == 0 && c.Beta == 0 && c.Gamma == 0 {
		return fmt.Errorf("candidate %q must enable at least one signal", c.Name)
	}
	return nil
}

// Sample represents a user/item pair used for evaluation.
type Sample struct {
	UserID      string
	Namespace   string
	HoldoutItem string
	EventTime   time.Time
	EventCount  int
}

// Result captures aggregate metrics for a single candidate.
type Result struct {
	Name           string
	Config         CandidateConfig
	Total          int
	Hits           int
	Failures       int
	SumRank        float64
	SumReciprocal  float64
	UniqueItemSeen int
	AvgListLength  float64

	HitRate  float64
	MRR      float64
	AvgRank  float64
	Coverage float64
}

// Harness coordinates sampling users and scoring candidate blends.
type Harness struct {
	Store      *store.Store
	BaseConfig algorithm.Config
	OrgID      uuid.UUID
	Namespace  string
	K          int
	Limit      int
	MinEvents  int
	Lookback   time.Duration
	Candidates []CandidateConfig
}

// Run executes the evaluation for each candidate and returns the aggregated metrics.
func (h *Harness) Run(ctx context.Context) ([]Result, error) {
	if err := h.validate(); err != nil {
		return nil, err
	}

	samples, err := h.loadSamples(ctx)
	if err != nil {
		return nil, err
	}
	if len(samples) == 0 {
		return nil, fmt.Errorf("no evaluation samples found for namespace %q", h.Namespace)
	}

	results := make([]Result, 0, len(h.Candidates))
	for _, cand := range h.Candidates {
		engine := algorithm.NewEngine(h.BaseConfig, h.Store, nil)
		res := evaluateCandidate(ctx, engine, samples, h.K, h.OrgID, cand)
		results = append(results, res)
	}
	return results, nil
}

func (h *Harness) validate() error {
	if h.Store == nil {
		return errors.New("store is required")
	}
	if h.OrgID == uuid.Nil {
		return errors.New("org id is required")
	}
	if h.Namespace == "" {
		return errors.New("namespace is required")
	}
	if h.K <= 0 {
		h.K = 20
	}
	if h.Limit <= 0 {
		h.Limit = 200
	}
	if h.MinEvents <= 0 {
		h.MinEvents = 5
	}
	if h.Lookback <= 0 {
		h.Lookback = 30 * 24 * time.Hour
	}
	if len(h.Candidates) == 0 {
		return errors.New("at least one candidate configuration is required")
	}
	for _, cand := range h.Candidates {
		if err := cand.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (h *Harness) loadSamples(ctx context.Context) ([]Sample, error) {
	const sampleSQL = `
WITH ranked AS (
	SELECT
		user_id,
		namespace,
		item_id,
		ts,
		ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY ts DESC) AS rnk,
		COUNT(*)    OVER (PARTITION BY user_id) AS total
	FROM events
	WHERE org_id   = $1
	  AND namespace = $2
	  AND ts       >= $3
)
SELECT user_id, namespace, item_id, ts, total
FROM ranked
WHERE rnk = 1 AND total >= $4
ORDER BY ts DESC
LIMIT $5;
`
	since := time.Now().Add(-h.Lookback)
	rows, err := h.Store.Pool.Query(ctx, sampleSQL, h.OrgID, h.Namespace, since, h.MinEvents, h.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Sample, 0, h.Limit)
	for rows.Next() {
		var s Sample
		if err := rows.Scan(&s.UserID, &s.Namespace, &s.HoldoutItem, &s.EventTime, &s.EventCount); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func evaluateCandidate(
	ctx context.Context,
	engine *algorithm.Engine,
	samples []Sample,
	k int,
	orgID uuid.UUID,
	cand CandidateConfig,
) Result {
	res := Result{
		Name:   cand.Name,
		Config: cand,
	}
	unique := make(map[string]struct{})
	var totalListLength float64

	weights := &algorithm.BlendWeights{
		Pop:  cand.Alpha,
		Cooc: cand.Beta,
		ALS:  cand.Gamma,
	}

	for _, sample := range samples {
		req := algorithm.Request{
			OrgID:     orgID,
			UserID:    sample.UserID,
			Namespace: sample.Namespace,
			K:         k,
			Blend:     weights,
		}
		resp, _, err := engine.Recommend(ctx, req)
		if err != nil {
			res.Failures++
			continue
		}
		res.Total++
		totalListLength += float64(len(resp.Items))

		rank := -1
		for idx, item := range resp.Items {
			unique[item.ItemID] = struct{}{}
			if item.ItemID == sample.HoldoutItem && rank == -1 {
				rank = idx + 1
			}
		}
		if rank > 0 {
			res.Hits++
			res.SumRank += float64(rank)
			res.SumReciprocal += 1.0 / float64(rank)
		}
	}

	res.UniqueItemSeen = len(unique)
	if res.Total > 0 {
		res.AvgListLength = totalListLength / float64(res.Total)
		res.HitRate = float64(res.Hits) / float64(res.Total)
		if res.Hits > 0 {
			res.AvgRank = res.SumRank / float64(res.Hits)
		}
		if res.SumReciprocal > 0 {
			res.MRR = res.SumReciprocal / float64(res.Total)
		}
		if k > 0 {
			res.Coverage = float64(res.UniqueItemSeen) / (float64(res.Total) * float64(k))
		}
	}
	return res
}
