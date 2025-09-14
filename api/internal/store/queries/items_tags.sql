-- name: items_tags
--
-- description: Get tags for specified item IDs.
--
-- inputs:
--  1. org_id (uuid)
--  2. namespace (text)
--  3. item_ids (text[])
--
-- outputs:
--   item_id (text),
--   tags (text[])
SELECT item_id,
    tags
FROM items
WHERE org_id = $1
    AND namespace = $2
    AND item_id = ANY($3::text []);