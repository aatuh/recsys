-- 061_interaction_events.sql
-- Purpose: store interaction/outcome logs (clicks, purchases, etc.) for eval.
-- Partitioned by occurred_at for scale.

CREATE TABLE IF NOT EXISTS interaction_events (
  id          bigserial PRIMARY KEY,
  tenant_id    uuid NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,

  occurred_at  timestamptz NOT NULL,

  request_id   uuid,

  user_key     bytea,
  session_id   text,

  surface      text,
  segment      text,

  event_type   interaction_type NOT NULL,
  item_id      text NOT NULL,
  position     integer,

  value        numeric(18,6),

  meta         jsonb
) PARTITION BY RANGE (occurred_at);

CREATE TABLE IF NOT EXISTS interaction_events_default
PARTITION OF interaction_events DEFAULT;

CREATE INDEX IF NOT EXISTS interaction_events_tenant_time_idx
ON interaction_events (tenant_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS interaction_events_req_idx
ON interaction_events (tenant_id, request_id)
WHERE request_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS interaction_events_user_time_idx
ON interaction_events (tenant_id, user_key, occurred_at DESC)
WHERE user_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS interaction_events_item_time_idx
ON interaction_events (tenant_id, item_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS interaction_events_meta_gin
ON interaction_events USING gin (meta);
