-- name: segments_upsert
INSERT INTO public.segments (
    org_id,
    namespace,
    segment_id,
    name,
    priority,
    active,
    profile_id,
    description
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) ON CONFLICT (org_id, namespace, segment_id)
DO UPDATE SET
    name = EXCLUDED.name,
    priority = EXCLUDED.priority,
    active = EXCLUDED.active,
    profile_id = EXCLUDED.profile_id,
    description = EXCLUDED.description,
    updated_at = now();
