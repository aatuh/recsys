-- 050_cache_invalidation.sql
-- Purpose: track admin cache invalidation requests and outcomes.

CREATE TABLE IF NOT EXISTS cache_invalidation_events (
  id               bigserial PRIMARY KEY,
  tenant_id        uuid NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  request_id       uuid,

  requested_at     timestamptz NOT NULL DEFAULT now(),
  requested_by_sub text NOT NULL,

  targets          text[] NOT NULL,
  surface          text,

  status           cache_invalidation_status NOT NULL DEFAULT 'requested',
  applied_at       timestamptz,
  applied_by       text,
  error_detail     text
);

CREATE INDEX IF NOT EXISTS cache_invalidation_tenant_time_idx
ON cache_invalidation_events (tenant_id, requested_at DESC);

CREATE INDEX IF NOT EXISTS cache_invalidation_status_idx
ON cache_invalidation_events (status);
