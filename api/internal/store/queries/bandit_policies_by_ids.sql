-- name: bandit_policies_by_ids
--
-- description: List bandit policies by IDs.
--
-- inputs:
--  1. org_id (text)
--  2. namespace (text)
--  3. policy_ids (text[])
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
    AND policy_id = ANY($3)
ORDER BY policy_id;