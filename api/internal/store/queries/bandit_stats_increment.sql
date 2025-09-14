-- name: bandit_stats_increment
--
-- description: Increment bandit statistics for a policy.
--
-- inputs:
--  1. org_id (text)
--  2. namespace (text)
--  3. surface (text)
--  4. bucket_key (text)
--  5. policy_id (text)
--  6. algo (text)
--  7. reward (boolean)
--
-- outputs: none (INSERT/UPDATE)
INSERT INTO bandit_stats (
        org_id,
        namespace,
        surface,
        bucket_key,
        policy_id,
        algo,
        trials,
        successes,
        alpha,
        beta,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        1,
        CASE
            WHEN $7 THEN 1
            ELSE 0
        END,
        1,
        1,
        NOW()
    ) ON CONFLICT (
        org_id,
        namespace,
        surface,
        bucket_key,
        policy_id,
        algo
    ) DO
UPDATE
SET trials = bandit_stats.trials + 1,
    successes = bandit_stats.successes + CASE
        WHEN EXCLUDED.successes = 1 THEN 1
        ELSE 0
    END,
    updated_at = NOW();