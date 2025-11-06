WITH tag_source AS (
    SELECT DISTINCT unnest($4::text[]) AS tag
),
ranked AS (
    SELECT i.item_id,
           COUNT(*)::float AS score
    FROM items i
    JOIN tag_source ts ON i.tags @> ARRAY[ts.tag]
    WHERE i.org_id = $1
      AND i.namespace = $2
      AND i.available = true
      AND NOT (i.item_id = ANY(COALESCE($5::text[], ARRAY[]::text[])))
    GROUP BY i.item_id
)
SELECT item_id, score
FROM ranked
ORDER BY score DESC, item_id
LIMIT $3;
