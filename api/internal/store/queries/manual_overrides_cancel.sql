UPDATE manual_overrides
SET status = 'cancelled',
    cancelled_at = now(),
    cancelled_by = $3
WHERE override_id = $1
  AND org_id = $2
  AND status = 'active'
RETURNING
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
    status,
    cancelled_at,
    cancelled_by;
