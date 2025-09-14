-- name: bandit_policies_active
--
-- description: List active bandit policies.
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
    AND is_active = true
ORDER BY policy_id;