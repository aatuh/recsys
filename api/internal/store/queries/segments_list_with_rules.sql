-- name: segments_list_with_rules
SELECT
    s.segment_id,
    s.name,
    s.priority,
    s.active,
    s.profile_id,
    s.description,
    s.created_at,
    s.updated_at,
    r.rule_id,
    r.rule,
    r.enabled,
    r.description,
    r.created_at,
    r.updated_at
FROM public.segments AS s
LEFT JOIN public.segment_rules AS r
  ON r.org_id = s.org_id
 AND r.namespace = s.namespace
 AND r.segment_id = s.segment_id
WHERE s.org_id = $1 AND s.namespace = $2
ORDER BY s.priority DESC, s.segment_id ASC, r.rule_id ASC;
