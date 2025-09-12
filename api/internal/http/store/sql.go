package store

// Rank items by time-decayed event activity, weighted by event type and
// half-life. Higher recent activity yields higher scores.
const popularitySQL = `
WITH w AS (
  SELECT
    COALESCE(tc.type, d.type) AS type,  -- event type
    COALESCE(tc.weight, d.weight) AS w, -- weight per type
    COALESCE(
      NULLIF(tc.half_life_days, 0),     -- per-type override
      NULLIF($3, 0),                    -- global override (days)
      d.half_life_days,                 -- default
      14                                -- fallback (days)
    ) AS hl
  FROM event_type_defaults d
  FULL OUTER JOIN event_type_config tc
    ON tc.org_id = $1 AND tc.namespace = $2 AND tc.type = d.type
  WHERE COALESCE(tc.is_active, true) = true
)
SELECT
  e.item_id,
  SUM(
    POWER(0.5, EXTRACT(EPOCH FROM ($10 - e.ts)) / (NULLIF(w.hl, 0) * 86400.0))
    * w.w * COALESCE(e.value, 1)
  ) AS score
FROM events e
JOIN w ON w.type = e.type
JOIN items i ON i.org_id = e.org_id AND i.namespace = e.namespace AND i.item_id = e.item_id
WHERE e.org_id = $1
  AND e.namespace = $2
  AND i.available = true
  AND ($5::timestamptz IS NULL OR e.ts >= $5)   -- optional earliest event ts
  AND ($6::float8     IS NULL OR i.price >= $6) -- optional min price
  AND ($7::float8     IS NULL OR i.price <= $7) -- optional max price
  AND (COALESCE($8::text[], '{}'::text[]) = '{}'::text[] OR i.tags && $8::text[])              -- optional tag overlap
  AND (COALESCE($9::text[], '{}'::text[]) = '{}'::text[] OR NOT (i.item_id = ANY($9::text[]))) -- optional exclusions
GROUP BY e.item_id
ORDER BY score DESC
LIMIT $4;
`
