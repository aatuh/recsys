SELECT
    override_id,
    org_id,
    namespace,
    surface,
    action,
    item_id,
    boost_value,
    notes,
    created_by,
    created_at,
    expires_at,
    rule_id,
    CASE
        WHEN status = 'active' AND expires_at IS NOT NULL AND expires_at <= now()
            THEN 'expired'
        ELSE status
    END AS status,
    cancelled_at,
    cancelled_by
FROM manual_overrides
WHERE org_id = $1
  AND ($2::text IS NULL OR namespace = $2)
  AND ($3::text IS NULL OR surface = $3)
  AND (
        $4::text IS NULL
        OR status = $4
        OR ($4 = 'expired' AND status = 'active' AND expires_at IS NOT NULL AND expires_at <= now())
      )
  AND ($5::text IS NULL OR action = $5)
ORDER BY created_at DESC;
