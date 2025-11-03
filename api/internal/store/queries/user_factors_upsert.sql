INSERT INTO recsys_user_factors (
    org_id,
    namespace,
    user_id,
    factors,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    CAST($4 AS vector(384)),
    now()
)
ON CONFLICT (org_id, namespace, user_id)
DO UPDATE
SET factors = EXCLUDED.factors,
    updated_at = now();
