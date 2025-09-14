package store

import (
	"context"
	"encoding/json"
	"strings"

	"recsys/internal/bandit"

	"github.com/jackc/pgx/v5"
)

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
	const q = `
INSERT INTO bandit_policies
(org_id, namespace, policy_id, name, is_active, config, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (org_id, namespace, policy_id) DO UPDATE
SET name = EXCLUDED.name,
    is_active = EXCLUDED.is_active,
    config = EXCLUDED.config,
    updated_at = NOW()
`
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
	const q = `
SELECT policy_id, name, is_active, config
FROM bandit_policies
WHERE org_id = $1 AND namespace = $2 AND is_active = true
ORDER BY policy_id
`
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
	const q = `
SELECT policy_id, name, is_active, config
FROM bandit_policies
WHERE org_id = $1 AND namespace = $2 AND policy_id = ANY($3)
ORDER BY policy_id
`
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
	const q = `
SELECT policy_id, trials, successes, alpha, beta
FROM bandit_stats
WHERE org_id = $1 AND namespace = $2
  AND surface = $3 AND bucket_key = $4 AND algo = $5
`
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
	const q = `
INSERT INTO bandit_stats
(org_id, namespace, surface, bucket_key, policy_id, algo,
 trials, successes, alpha, beta, updated_at)
VALUES ($1,$2,$3,$4,$5,$6, 1, CASE WHEN $7 THEN 1 ELSE 0 END, 1, 1, NOW())
ON CONFLICT (org_id, namespace, surface, bucket_key, policy_id, algo)
DO UPDATE SET
  trials = bandit_stats.trials + 1,
  successes = bandit_stats.successes + CASE WHEN EXCLUDED.successes=1 THEN 1 ELSE 0 END,
  updated_at = NOW()
`
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
	const q = `
INSERT INTO bandit_decisions_log
(org_id, namespace, surface, bucket_key, policy_id, algo, explore, request_id, meta)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
`
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
	const q = `
INSERT INTO bandit_rewards_log
(org_id, namespace, surface, bucket_key, policy_id, algo, reward, request_id, meta)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
`
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
