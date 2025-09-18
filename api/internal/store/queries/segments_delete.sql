-- name: segments_delete
DELETE FROM public.segments
WHERE org_id = $1 AND namespace = $2 AND segment_id = ANY($3::text[]);
