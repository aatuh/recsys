-- name: user_purchased_since
--
-- description: Get distinct item IDs purchased by user since timestamp (type=3).
--
-- inputs:
--  1. org_id (uuid)
--  2. namespace (text)
--  3. user_id (text)
--  4. since (timestamptz)
--
-- outputs:
--   item_id (text)
SELECT DISTINCT item_id
FROM events
WHERE org_id = $1
    AND namespace = $2
    AND user_id = $3
    AND type = 3
    AND ts >= $4;