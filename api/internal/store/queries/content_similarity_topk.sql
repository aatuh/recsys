WITH tag_source AS (
    SELECT unnest($4::text[]) AS tag
),
ranked AS (
    SELECT it.item_id,
           COUNT(*)::float AS score
    FROM items_tags it
    JOIN tag_source ts ON it.tag = ts.tag
    WHERE it.org_id = $1
      AND it.namespace = $2
      AND NOT (it.item_id = ANY(COALESCE($5::text[], ARRAY[]::text[])))
    GROUP BY it.item_id
)
SELECT item_id, score
FROM ranked
ORDER BY score DESC, item_id
LIMIT $3;
