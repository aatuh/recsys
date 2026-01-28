-- 060_exposure_events.sql
-- Purpose: store exposure logs for evaluation and debugging.
-- Partitioned by occurred_at for scale.

CREATE TABLE IF NOT EXISTS exposure_events (
  id                bigserial PRIMARY KEY,
  tenant_id          uuid NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,

  occurred_at        timestamptz NOT NULL,
  request_id         uuid NOT NULL,

  surface            text NOT NULL,
  segment            text NOT NULL,

  user_key           bytea,
  session_id         text,

  experiment_id      text,
  experiment_variant text,

  algo_version       text NOT NULL,
  config_etag        text,
  rules_etag         text,

  request            jsonb NOT NULL,
  response           jsonb NOT NULL,
  meta               jsonb
) PARTITION BY RANGE (occurred_at);

-- Default partition prevents insert failures when a time partition is missing.
CREATE TABLE IF NOT EXISTS exposure_events_default
PARTITION OF exposure_events DEFAULT;

-- Indexes (partitioned indexes: created per partition).
CREATE INDEX IF NOT EXISTS exposure_events_tenant_time_idx
ON exposure_events (tenant_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS exposure_events_req_idx
ON exposure_events (tenant_id, request_id);

CREATE INDEX IF NOT EXISTS exposure_events_surface_time_idx
ON exposure_events (tenant_id, surface, occurred_at DESC);

CREATE INDEX IF NOT EXISTS exposure_events_user_time_idx
ON exposure_events (tenant_id, user_key, occurred_at DESC)
WHERE user_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS exposure_events_request_gin
ON exposure_events USING gin (request);

CREATE INDEX IF NOT EXISTS exposure_events_response_gin
ON exposure_events USING gin (response);
