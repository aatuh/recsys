-- 090_signal_tables.sql
-- Purpose: store item tags and daily popularity signals.

CREATE TABLE IF NOT EXISTS item_tags (
  tenant_id  uuid NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  namespace  text NOT NULL DEFAULT 'default',
  item_id    text NOT NULL,
  tags       text[] NOT NULL DEFAULT '{}',
  price      numeric(18,6),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, namespace, item_id)
);

DROP TRIGGER IF EXISTS item_tags_set_updated_at ON item_tags;
CREATE TRIGGER item_tags_set_updated_at
BEFORE UPDATE ON item_tags
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS item_tags_tags_gin
ON item_tags USING gin (tags);

CREATE INDEX IF NOT EXISTS item_tags_tenant_ns_created_idx
ON item_tags (tenant_id, namespace, created_at DESC);

CREATE TABLE IF NOT EXISTS item_popularity_daily (
  tenant_id  uuid NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  namespace  text NOT NULL DEFAULT 'default',
  item_id    text NOT NULL,
  day        date NOT NULL,
  score      numeric(18,6) NOT NULL DEFAULT 0,
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, namespace, item_id, day)
);

DROP TRIGGER IF EXISTS item_popularity_daily_set_updated_at ON item_popularity_daily;
CREATE TRIGGER item_popularity_daily_set_updated_at
BEFORE UPDATE ON item_popularity_daily
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS item_popularity_daily_tenant_ns_day_idx
ON item_popularity_daily (tenant_id, namespace, day DESC);

CREATE INDEX IF NOT EXISTS item_popularity_daily_tenant_item_day_idx
ON item_popularity_daily (tenant_id, item_id, day DESC);
