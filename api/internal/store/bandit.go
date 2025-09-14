package store

import (
	"context"
	"encoding/json"
	"strings"

	"recsys/internal/bandit"

	_ "embed"

	"github.com/jackc/pgx/v5"
)

//go:embed queries/bandit_policies_upsert.sql
var banditPoliciesUpsertSQL string

//go:embed queries/bandit_policies_active.sql
var banditPoliciesActiveSQL string

//go:embed queries/bandit_policies_by_ids.sql
var banditPoliciesByIdsSQL string

//go:embed queries/bandit_stats_get.sql
var banditStatsGetSQL string

//go:embed queries/bandit_stats_increment.sql
var banditStatsIncrementSQL string

//go:embed queries/bandit_decisions_log.sql
var banditDecisionsLogSQL string

//go:embed queries/bandit_rewards_log.sql
var banditRewardsLogSQL string

type BanditPolicyRow struct {
	PolicyID   string
	Namespace  string
	Name       string
	IsActive   bool
	ConfigJSON []byte
}

func (s *Store) UpsertBanditPolicies(
	ctx context.Context, orgID string, ns string, rows []bandit.PolicyConfig,
) error {
	if len(rows) == 0 {
		return nil
	}
	q := banditPoliciesUpsertSQL
	batch := &pgx.Batch{}
	for _, r := range rows {
		cfg, _ := json.Marshal(r)
		batch.Queue(q, orgID, ns, r.PolicyID, r.Name, r.Active, cfg)
	}
	br := s.Pool.SendBatch(ctx, batch)
	defer br.Close()
	for range rows {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ListActivePolicies(
	ctx context.Context, orgID string, ns string,
) ([]bandit.PolicyConfig, error) {
	q := banditPoliciesActiveSQL
	rows, err := s.Pool.Query(ctx, q, orgID, ns)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPolicies(rows)
}

func (s *Store) ListPoliciesByIDs(
	ctx context.Context, orgID, ns string, ids []string,
) ([]bandit.PolicyConfig, error) {
	if len(ids) == 0 {
		return []bandit.PolicyConfig{}, nil
	}
	q := banditPoliciesByIdsSQL
	rows, err := s.Pool.Query(ctx, q, orgID, ns, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPolicies(rows)
}

func scanPolicies(rows pgx.Rows) ([]bandit.PolicyConfig, error) {
	var out []bandit.PolicyConfig
	for rows.Next() {
		var id, name string
		var active bool
		var cfg []byte
		if err := rows.Scan(&id, &name, &active, &cfg); err != nil {
			return nil, err
		}
		var pc bandit.PolicyConfig
		if err := json.Unmarshal(cfg, &pc); err != nil {
			return nil, err
		}
		// Ensure identity matches row key.
		pc.PolicyID = id
		pc.Name = name
		pc.Active = active
		out = append(out, pc)
	}
	return out, rows.Err()
}

func (s *Store) GetStats(
	ctx context.Context,
	orgID, ns, surface, bucket string,
	algo bandit.Algorithm,
) (map[string]bandit.Stats, error) {
	q := banditStatsGetSQL
	rows, err := s.Pool.Query(ctx, q, orgID, ns, surface, bucket, string(algo))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bandit.Stats{}
	for rows.Next() {
		var pid string
		var st bandit.Stats
		if err := rows.Scan(&pid, &st.Trials, &st.Successes, &st.Alpha, &st.Beta); err != nil {
			return nil, err
		}
		out[pid] = st
	}
	return out, rows.Err()
}

func (s *Store) IncrementStats(
	ctx context.Context,
	orgID, ns, surface, bucket string,
	algo bandit.Algorithm, policyID string, reward bool,
) error {
	q := banditStatsIncrementSQL
	_, err := s.Pool.Exec(ctx, q, orgID, ns, surface, bucket, policyID,
		string(algo), reward)
	return err
}

func (s *Store) LogDecision(
	ctx context.Context,
	orgID, ns, surface, bucket string,
	algo bandit.Algorithm, policyID string, explore bool,
	reqID string, meta map[string]any,
) error {
	q := banditDecisionsLogSQL
	var js []byte
	if meta != nil {
		js, _ = json.Marshal(meta)
	}
	_, err := s.Pool.Exec(ctx, q, orgID, ns, surface, bucket, policyID,
		string(algo), explore, nullIfEmpty(reqID), js)
	return err
}

func (s *Store) LogReward(
	ctx context.Context,
	orgID, ns, surface, bucket string,
	algo bandit.Algorithm, policyID string, reward bool,
	reqID string, meta map[string]any,
) error {
	q := banditRewardsLogSQL
	var js []byte
	if meta != nil {
		js, _ = json.Marshal(meta)
	}
	_, err := s.Pool.Exec(ctx, q, orgID, ns, surface, bucket, policyID,
		string(algo), reward, nullIfEmpty(reqID), js)
	return err
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
