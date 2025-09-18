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
WHERE org_id = $1 AND rule_id = $2;
