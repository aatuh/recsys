-- name: bandit_rewards_log
--
-- description: Log bandit reward.
--
-- inputs:
--  1. org_id (text)
--  2. namespace (text)
--  3. surface (text)
--  4. bucket_key (text)
--  5. policy_id (text)
--  6. algo (text)
--  7. reward (boolean)
--  8. request_id (text|null)
--  9. meta (jsonb)
--
-- outputs: none (INSERT)
INSERT INTO bandit_rewards_log (
        org_id,
        namespace,
        surface,
        bucket_key,
        policy_id,
        algo,
        reward,
        request_id,
        meta
    )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);