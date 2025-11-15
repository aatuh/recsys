SELECT r.rule_id,
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
  AND (
    $2::text IS NULL
    OR r.namespace = $2::text
  )
  AND (
    $3::text IS NULL
    OR r.surface = $3::text
  )
  AND (
    $4::text IS NULL
    OR (
      (
        $4::text = '__NULL__'
        AND segment_id IS NULL
      )
      OR (r.segment_id = $4::text)
    )
  )
  AND (
    $5::boolean IS NULL
    OR r.enabled = $5::boolean
  )
  AND (
    $6::timestamptz IS NULL
    OR (
      (
        r.valid_from IS NULL
        OR r.valid_from <= $6::timestamptz
      )
      AND (
        r.valid_until IS NULL
        OR r.valid_until >= $6::timestamptz
      )
    )
  )
  AND (
    $7::text IS NULL
    OR r.action::text = $7::text
  )
  AND (
    $8::text IS NULL
    OR r.target_type::text = $8::text
  )
ORDER BY r.priority DESC,
  r.created_at ASC;
