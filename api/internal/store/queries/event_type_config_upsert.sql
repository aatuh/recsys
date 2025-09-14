-- name: event_type_config_upsert
--
-- description: Insert or update event type configuration.
--
-- inputs:
--  1. org_id (uuid)
--  2. namespace (text)
--  3. type (int2)
--  4. name (text|null)
--  5. weight (float8)
--  6. half_life_days (float8|null)
--  7. is_active (boolean|null)
--
-- outputs: none (INSERT/UPDATE)
INSERT INTO event_type_config (
        org_id,
        namespace,
        type,
        name,
        weight,
        half_life_days,
        is_active,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        COALESCE($7, true),
        now()
    ) ON CONFLICT (org_id, namespace, type) DO
UPDATE
SET name = COALESCE(EXCLUDED.name, event_type_config.name),
    weight = EXCLUDED.weight,
    half_life_days = EXCLUDED.half_life_days,
    is_active = COALESCE(EXCLUDED.is_active, event_type_config.is_active),
    updated_at = now();