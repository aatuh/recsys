-- name: items_upsert
--
-- description: Insert or update items with optional embedding vector.
--
-- inputs:
--  1. org_id (uuid)
--  2. namespace (text)
--  3. item_id (text)
--  4. available (boolean)
--  5. price (float8|null)
--  6. tags (text[])
--  7. props (jsonb)
--  8. embedding (text|null) - vector literal like "[0.1,0.2,...]"
--
-- outputs: none (INSERT/UPDATE)
INSERT INTO items (
        org_id,
        namespace,
        item_id,
        available,
        price,
        tags,
        props,
        embedding,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        COALESCE($6, '{}'::text []),
        COALESCE($7, '{}'::jsonb),
        CASE
            WHEN $8::text IS NULL THEN NULL
            ELSE CAST($8 AS vector(384))
        END,
        now(),
        now()
    ) ON CONFLICT (org_id, namespace, item_id) DO
UPDATE
SET available = EXCLUDED.available,
    price = COALESCE(EXCLUDED.price, items.price),
    tags = COALESCE(EXCLUDED.tags, items.tags),
    props = COALESCE(EXCLUDED.props, items.props),
    embedding = CASE
        WHEN EXCLUDED.embedding IS NULL THEN items.embedding
        ELSE EXCLUDED.embedding
    END,
    updated_at = now();