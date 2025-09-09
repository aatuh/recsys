package store

const popularitySQL = `
WITH w AS (
  SELECT
    COALESCE(tc.type, d.type) AS type,
    COALESCE(tc.weight, d.weight) AS w,
    COALESCE(
      NULLIF(tc.half_life_days, 0),
      NULLIF($3, 0),
      d.half_life_days,
      14
    ) AS hl
  FROM event_type_defaults d
  FULL OUTER JOIN event_type_config tc
    ON tc.org_id = $1 AND tc.namespace = $2 AND tc.type = d.type
  WHERE COALESCE(tc.is_active, true) = true
)
SELECT e.item_id,
       SUM(
         POWER(0.5, EXTRACT(EPOCH FROM (now() - e.ts)) / (NULLIF(w.hl,0)*86400.0))
         * w.w * COALESCE(e.value, 1)
       ) AS score
FROM events e
JOIN w ON w.type = e.type
JOIN items i ON i.org_id = e.org_id AND i.namespace = e.namespace AND i.item_id = e.item_id
WHERE e.org_id = $1
  AND e.namespace = $2
  AND i.available = true
  AND ($5::timestamptz IS NULL OR e.ts >= $5)
  AND ($6::float8     IS NULL OR i.price >= $6)
  AND ($7::float8     IS NULL OR i.price <= $7)
  AND (COALESCE($8::text[], '{}'::text[]) = '{}'::text[] OR i.tags && $8::text[])
  AND (COALESCE($9::text[], '{}'::text[]) = '{}'::text[] OR NOT (i.item_id = ANY($9::text[])))
GROUP BY e.item_id
ORDER BY score DESC
LIMIT $4;

`
