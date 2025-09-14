-- name: users_upsert
--
-- description: Insert or update users with traits.
--
-- inputs:
--  1. org_id (uuid)
--  2. namespace (text)
--  3. user_id (text)
--  4. traits (jsonb)
--
-- outputs: none (INSERT/UPDATE)
INSERT INTO users (
        org_id,
        namespace,
        user_id,
        traits,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        COALESCE($4, '{}'::jsonb),
        now(),
        now()
    ) ON CONFLICT (org_id, namespace, user_id) DO
UPDATE
SET traits = COALESCE(EXCLUDED.traits, users.traits),
    updated_at = now();