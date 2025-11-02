WITH user_vec AS (
    SELECT factors
    FROM recsys_user_factors
    WHERE org_id = $1
      AND namespace = $2
      AND user_id = $3
      AND factors IS NOT NULL
),
 scored AS (
    SELECT 
        items.item_id,
        (items.factors <#> uv.factors) * -1 AS score
    FROM recsys_item_factors items
    CROSS JOIN user_vec uv
    WHERE items.org_id = $1
      AND items.namespace = $2
      AND NOT (items.item_id = ANY(COALESCE($5::text[], ARRAY[]::text[])))
)
SELECT item_id, score
FROM scored
ORDER BY score DESC
LIMIT $4;
