-- name: bandit_stats_get
--
-- description: Get bandit statistics for policies.
--
-- inputs:
--  1. org_id (text)
--  2. namespace (text)
--  3. surface (text)
--  4. bucket_key (text)
--  5. algo (text)
--
-- outputs:
--   policy_id (text),
--   trials (int8),
--   successes (int8),
--   alpha (float8),
--   beta (float8)
SELECT policy_id,
    trials,
    successes,
    alpha,
    beta
FROM bandit_stats
WHERE org_id = $1
    AND namespace = $2
    AND surface = $3
    AND bucket_key = $4
    AND algo = $5;