-- 100_item_covisit_daily.sql
-- Purpose: store daily co-visitation counts for similar-items signals.

CREATE TABLE IF NOT EXISTS item_covisit_daily (
  tenant_id   uuid NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  namespace   text NOT NULL DEFAULT 'default',
  item_id     text NOT NULL,
  neighbor_id text NOT NULL,
  day         date NOT NULL,
  score       numeric(18,6) NOT NULL DEFAULT 0,
  updated_at  timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, namespace, item_id, neighbor_id, day)
);

DROP TRIGGER IF EXISTS item_covisit_daily_set_updated_at ON item_covisit_daily;
CREATE TRIGGER item_covisit_daily_set_updated_at
BEFORE UPDATE ON item_covisit_daily
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS item_covisit_daily_tenant_item_day_idx
ON item_covisit_daily (tenant_id, namespace, item_id, day DESC);

CREATE INDEX IF NOT EXISTS item_covisit_daily_tenant_neighbor_day_idx
ON item_covisit_daily (tenant_id, namespace, neighbor_id, day DESC);
