-- name: embedding_similarity
--
-- description: Find similar items by embedding cosine distance.
--
-- inputs:
--  1. org_id (uuid)
--  2. namespace (text)
--  3. item_id (text)
--  4. k (int)
--
-- outputs:
--   item_id (text),
--   score (float8) - similarity score (1 - distance)
WITH anchor AS (
    SELECT embedding
    FROM items
    WHERE org_id = $1
        AND namespace = $2
        AND item_id = $3
        AND embedding IS NOT NULL
)
SELECT i.item_id,
    (1.0 - (a.embedding <=> i.embedding)) AS score
FROM anchor a
    JOIN items i ON i.org_id = $1
    AND i.namespace = $2
WHERE i.item_id <> $3
    AND i.available = true
    AND i.embedding IS NOT NULL
ORDER BY a.embedding <=> i.embedding ASC
LIMIT $4