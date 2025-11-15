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
WHERE r.org_id = $1 AND r.rule_id = $2;
