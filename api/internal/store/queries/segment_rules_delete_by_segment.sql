-- name: segment_rules_delete_by_segment
DELETE FROM public.segment_rules
WHERE org_id = $1 AND namespace = $2 AND segment_id = $3;
