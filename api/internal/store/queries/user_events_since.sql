-- name: user_events_since
--
-- description: Get distinct item IDs for specified user event types since timestamp.
--
-- inputs:
--  1. org_id (uuid)
--  2. namespace (text)
--  3. user_id (text)
--  4. since (timestamptz)
--  5. event_types (smallint[] | null for any)
--
-- outputs:
--   item_id (text)
SELECT DISTINCT item_id
FROM events
WHERE org_id = $1
    AND namespace = $2
    AND user_id = $3
    AND ts >= $4
    AND (
        $5::smallint[] IS NULL
        OR cardinality($5::smallint[]) = 0
        OR type = ANY($5::smallint[])
    );
