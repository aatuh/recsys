WITH recent_events AS (
    SELECT item_id, ts
    FROM events
    WHERE org_id = $1
      AND namespace = $2
      AND user_id = $3
      AND item_id IS NOT NULL
    ORDER BY ts DESC
    LIMIT $4
),
cooccurrence AS (
    SELECT e2.item_id,
           COUNT(*) AS score
    FROM recent_events r
    JOIN events e2
      ON e2.org_id = $1
     AND e2.namespace = $2
     AND e2.user_id = $3
     AND e2.item_id IS NOT NULL
     AND e2.item_id <> r.item_id
    WHERE e2.ts > r.ts
      AND e2.ts <= r.ts + ($5 * interval '1 minute')
    GROUP BY e2.item_id
)
SELECT item_id, score
FROM cooccurrence
WHERE NOT (item_id = ANY(COALESCE($6::text[], ARRAY[]::text[])))
ORDER BY score DESC, item_id
LIMIT $7;
