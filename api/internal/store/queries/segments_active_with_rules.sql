-- name: segments_active_with_rules
SELECT
    s.segment_id,
    s.name,
    s.priority,
    s.profile_id,
    s.description,
    r.rule_id,
    r.rule,
    r.enabled
FROM public.segments AS s
LEFT JOIN public.segment_rules AS r
  ON r.org_id = s.org_id
 AND r.namespace = s.namespace
 AND r.segment_id = s.segment_id
WHERE s.org_id = $1
  AND s.namespace = $2
  AND s.active = true
ORDER BY s.priority DESC, s.segment_id ASC, r.rule_id ASC;
