-- name: cooccurrence_top_k
--
-- description: Find co-occurrence neighbors for an item within a time window.
--
-- inputs:
--  1. org_id (uuid)
--  2. namespace (text)
--  3. item_id (text)
--  4. k (int)
--  5. since (timestamptz)
--
-- outputs:
--   item_id (text),
--   score (float8) - co-occurrence count
SELECT e2.item_id,
    COUNT(*)::float8 AS c
FROM events e1
    JOIN events e2 ON e1.org_id = e2.org_id
    AND e1.namespace = e2.namespace
    AND e1.user_id = e2.user_id
    AND e2.item_id <> $3
WHERE e1.org_id = $1
    AND e1.namespace = $2
    AND e1.item_id = $3
    AND e1.ts > $5
    AND e2.ts > $5
GROUP BY e2.item_id
ORDER BY c DESC
LIMIT $4;