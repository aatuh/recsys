-- name: segment_rules_insert
INSERT INTO public.segment_rules (
    org_id,
    namespace,
    segment_id,
    rule,
    enabled,
    description
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING rule_id;
