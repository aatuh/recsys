-- name: event_type_config_effective
--
-- description: Get effective event type configuration (tenant override if exists, else default).
--
-- inputs:
--  1. org_id (uuid)
--  2. namespace (text)
--
-- outputs:
--   type (int2),
--   name (text),
--   weight (float8),
--   half_life_days (float8),
--   is_active (boolean),
--   source (text) - "tenant" or "default"
SELECT COALESCE(tc.type, d.type) AS type,
    COALESCE(tc.name, d.name) AS name,
    COALESCE(tc.weight, d.weight) AS weight,
    COALESCE(tc.half_life_days, d.half_life_days) AS half_life_days,
    COALESCE(tc.is_active, true) AS is_active,
    CASE
        WHEN tc.type IS NULL THEN 'default'
        ELSE 'tenant'
    END AS source
FROM event_type_defaults d
    FULL OUTER JOIN event_type_config tc ON tc.org_id = $1
    AND tc.namespace = $2
    AND tc.type = d.type
WHERE COALESCE(tc.is_active, true) = true
ORDER BY type;