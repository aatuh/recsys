-- name: events_insert
--
-- description: Insert events with optional source event ID for deduplication.
--
-- inputs:
--  1. org_id (uuid)
--  2. namespace (text)
--  3. user_id (text)
--  4. item_id (text)
--  5. type (int2)
--  6. value (float8)
--  7. ts (timestamptz)
--  8. meta (jsonb)
--  9. source_event_id (text|null)
--
-- outputs: none (INSERT)
INSERT INTO events (
        org_id,
        namespace,
        user_id,
        item_id,
        type,
        value,
        ts,
        meta,
        source_event_id
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        COALESCE($8, '{}'::jsonb),
        $9
    ) ON CONFLICT (org_id, namespace, source_event_id) DO NOTHING;