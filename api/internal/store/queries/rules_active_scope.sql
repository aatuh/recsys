SELECT
    r.rule_id,
    r.org_id,
    r.namespace,
    r.surface,
    r.name,
    r.description,
    r.action,
    r.target_type,
    r.target_key,
    r.item_ids,
    r.boost_value,
    r.max_pins,
    r.segment_id,
    r.priority,
    r.enabled,
    r.valid_from,
    r.valid_until,
    r.created_at,
    r.updated_at,
    mo.override_id
FROM rules r
LEFT JOIN manual_overrides mo
  ON mo.rule_id = r.rule_id
  AND mo.status = 'active'
WHERE r.org_id = $1
  AND r.namespace = $2
  AND r.surface = $3
  AND r.enabled = true
  AND (r.segment_id IS NULL OR r.segment_id = $4)
  AND ($5 = false OR r.valid_from IS NULL OR r.valid_from <= $6)
  AND ($5 = false OR r.valid_until IS NULL OR r.valid_until >= $6)
ORDER BY r.priority DESC, r.created_at ASC;
