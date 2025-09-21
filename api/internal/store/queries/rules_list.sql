SELECT rule_id,
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
  AND (
    $2::text IS NULL
    OR namespace = $2::text
  )
  AND (
    $3::text IS NULL
    OR surface = $3::text
  )
  AND (
    $4::text IS NULL
    OR (
      (
        $4::text = '__NULL__'
        AND segment_id IS NULL
      )
      OR (segment_id = $4::text)
    )
  )
  AND (
    $5::boolean IS NULL
    OR enabled = $5::boolean
  )
  AND (
    $6::timestamptz IS NULL
    OR (
      (
        valid_from IS NULL
        OR valid_from <= $6::timestamptz
      )
      AND (
        valid_until IS NULL
        OR valid_until >= $6::timestamptz
      )
    )
  )
  AND (
    $7::text IS NULL
    OR action::text = $7::text
  )
  AND (
    $8::text IS NULL
    OR target_type::text = $8::text
  )
ORDER BY priority DESC,
  created_at ASC;