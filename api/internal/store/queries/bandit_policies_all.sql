-- name: bandit_policies_all
--
-- description: List all bandit policies (active and inactive).
--
-- inputs:
--  1. org_id (text)
--  2. namespace (text)
--
-- outputs:
--   policy_id (text),
--   name (text),
--   is_active (boolean),
--   config (jsonb)
SELECT policy_id,
    name,
    is_active,
    config
FROM bandit_policies
WHERE org_id = $1
    AND namespace = $2
ORDER BY is_active DESC,
    policy_id;