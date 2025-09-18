-- name: user_get
SELECT user_id, traits, created_at, updated_at
FROM public.users
WHERE org_id = $1 AND namespace = $2 AND user_id = $3;
