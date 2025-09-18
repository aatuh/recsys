SELECT
    rule_id,
    org_id,
    namespace,
    surface,
    name,
    description,
    action,
    target_type,
    target_key,
    item_ids,
    boost_value,
    max_pins,
    segment_id,
    priority,
    enabled,
    valid_from,
    valid_until,
    created_at,
    updated_at
FROM rules
WHERE org_id = $1
  AND namespace = $2
  AND surface = $3
  AND enabled = true
  AND (segment_id IS NULL OR segment_id = $4)
  AND ($5 = false OR valid_from IS NULL OR valid_from <= $6)
  AND ($5 = false OR valid_until IS NULL OR valid_until >= $6)
ORDER BY priority DESC, created_at ASC;
