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
--  9. brand (text|null)
-- 10. category (text|null)
-- 11. category_path (text[]|null)
-- 12. description (text|null)
-- 13. image_url (text|null)
-- 14. metadata_version (text|null)
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
        brand,
        category,
        category_path,
        description,
        image_url,
        metadata_version,
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
        $9,
        $10,
        COALESCE($11::text [], '{}'::text []),
        $12,
        $13,
        $14,
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
    brand = COALESCE(EXCLUDED.brand, items.brand),
    category = COALESCE(EXCLUDED.category, items.category),
    category_path = COALESCE(EXCLUDED.category_path, items.category_path),
    description = COALESCE(EXCLUDED.description, items.description),
    image_url = COALESCE(EXCLUDED.image_url, items.image_url),
    metadata_version = COALESCE(EXCLUDED.metadata_version, items.metadata_version),
    updated_at = now();
