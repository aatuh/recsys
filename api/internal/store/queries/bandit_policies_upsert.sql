-- name: bandit_policies_upsert
--
-- description: Insert or update bandit policies.
--
-- inputs:
--  1. org_id (text)
--  2. namespace (text)
--  3. policy_id (text)
--  4. name (text)
--  5. is_active (boolean)
--  6. config (jsonb)
--
-- outputs: none (INSERT/UPDATE)
INSERT INTO bandit_policies (
        org_id,
        namespace,
        policy_id,
        name,
        is_active,
        config,
        updated_at
    )
VALUES ($1, $2, $3, $4, $5, $6, NOW()) ON CONFLICT (org_id, namespace, policy_id) DO
UPDATE
SET name = EXCLUDED.name,
    is_active = EXCLUDED.is_active,
    config = EXCLUDED.config,
    updated_at = NOW();