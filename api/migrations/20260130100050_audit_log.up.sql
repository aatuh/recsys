-- 040_audit_log.sql
-- Purpose: append-only audit trail for admin/control-plane actions.

CREATE TABLE IF NOT EXISTS audit_log (
  id           bigserial PRIMARY KEY,
  occurred_at  timestamptz NOT NULL DEFAULT now(),
  tenant_id    uuid REFERENCES tenants(id) ON DELETE SET NULL,

  actor_sub    text NOT NULL,
  actor_type   text NOT NULL,
  action       text NOT NULL,

  entity_type  text,
  entity_id    text,

  request_id   uuid,
  ip           inet,
  user_agent   text,

  before_state jsonb,
  after_state  jsonb,
  extra        jsonb
);

DROP TRIGGER IF EXISTS audit_log_immutable_ud ON audit_log;
CREATE TRIGGER audit_log_immutable_ud
BEFORE UPDATE OR DELETE ON audit_log
FOR EACH ROW EXECUTE FUNCTION prevent_update_delete();

CREATE INDEX IF NOT EXISTS audit_log_tenant_time_idx
ON audit_log (tenant_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS audit_log_action_time_idx
ON audit_log (action, occurred_at DESC);

CREATE INDEX IF NOT EXISTS audit_log_request_id_idx
ON audit_log (request_id);
