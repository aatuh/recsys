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
  AND ($2 IS NULL OR namespace = $2)
  AND ($3 IS NULL OR surface = $3)
  AND (
        $4 IS NULL OR (
          ($4 = '__NULL__' AND segment_id IS NULL) OR
          (segment_id = $4)
        )
      )
  AND ($5 IS NULL OR enabled = $5)
  AND (
        $6 IS NULL OR (
          (valid_from IS NULL OR valid_from <= $6)
          AND (valid_until IS NULL OR valid_until >= $6)
        )
      )
  AND ($7 IS NULL OR action = $7::rule_action)
  AND ($8 IS NULL OR target_type = $8::rule_target)
ORDER BY priority DESC, created_at ASC;
