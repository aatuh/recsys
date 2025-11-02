UPDATE manual_overrides
SET status = 'expired'
WHERE org_id = $1
  AND status = 'active'
  AND expires_at IS NOT NULL
  AND expires_at <= $2;
