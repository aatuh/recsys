WITH filtered AS (
    SELECT item_id,
           available,
           price,
           tags,
           props,
           embedding,
           brand,
           category,
           category_path,
           description,
           image_url,
           metadata_version,
           updated_at
    FROM items
    WHERE org_id = $1
      AND namespace = $2
      AND (
            NOT $3::boolean
            OR brand IS NULL
            OR category IS NULL
            OR description IS NULL
            OR image_url IS NULL
            OR metadata_version IS NULL
            OR COALESCE(array_length(category_path, 1), 0) = 0
      )
      AND (
            $4::timestamptz IS NULL
            OR updated_at >= $4
      )
)
SELECT item_id,
       available,
       price,
       tags,
       props,
       embedding,
       brand,
       category,
       category_path,
       description,
       image_url,
       metadata_version,
       updated_at
FROM filtered
WHERE (
        $5::timestamptz IS NULL
        OR updated_at < $5
        OR (
            updated_at = $5
            AND (
                $6::text IS NULL
                OR item_id < $6
            )
        )
)
ORDER BY updated_at DESC, item_id DESC
LIMIT $7;
