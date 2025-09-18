-- name: segment_profiles_delete
DELETE FROM public.segment_profiles
WHERE org_id = $1 AND namespace = $2 AND profile_id = ANY($3::text[]);
