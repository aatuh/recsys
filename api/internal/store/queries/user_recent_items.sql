-- name: user_recent_items
--
-- description: Get recent items for user since timestamp, ordered by most recent interaction.
--
-- inputs:
--  1. org_id (uuid)
--  2. namespace (text)
--  3. user_id (text)
--  4. since (timestamptz)
--  5. limit (int)
--
-- outputs:
--   item_id (text)
SELECT item_id
FROM (
        SELECT item_id,
            MAX(ts) AS last_ts
        FROM events
        WHERE org_id = $1
            AND namespace = $2
            AND user_id = $3
            AND ts >= $4
        GROUP BY item_id
    ) u
ORDER BY last_ts DESC
LIMIT $5;