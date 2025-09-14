-- name: user_tag_profile
--
-- description: Build user tag profile from decayed, weighted event activity.
--
-- inputs:
--  1. org_id (uuid)
--  2. namespace (text)
--  3. user_id (text)
--  4. since (timestamptz|null)
--  5. top_n (int)
--
-- outputs:
--   tag (text),
--   score (float8) - decayed weighted score
WITH w AS (
    SELECT COALESCE(tc.type, d.type) AS type,
        COALESCE(tc.weight, d.weight) AS w,
        COALESCE(
            NULLIF(tc.half_life_days, 0),
            d.half_life_days,
            14
        ) AS hl
    FROM event_type_defaults d
        FULL OUTER JOIN event_type_config tc ON tc.org_id = $1
        AND tc.namespace = $2
        AND tc.type = d.type
    WHERE COALESCE(tc.is_active, true) = true
)
SELECT t.tag,
    SUM(
        POWER(
            0.5,
            EXTRACT(
                EPOCH
                FROM (now() - e.ts)
            ) / (NULLIF(w.hl, 0) * 86400.0)
        ) * w.w * COALESCE(e.value, 1)
    ) AS s
FROM events e
    JOIN w ON w.type = e.type
    JOIN items i ON i.org_id = e.org_id
    AND i.namespace = e.namespace
    AND i.item_id = e.item_id
    CROSS JOIN LATERAL UNNEST(i.tags) AS t(tag)
WHERE e.org_id = $1
    AND e.namespace = $2
    AND e.user_id = $3
    AND (
        $4::timestamptz IS NULL
        OR e.ts >= $4
    )
GROUP BY t.tag
ORDER BY s DESC
LIMIT $5;