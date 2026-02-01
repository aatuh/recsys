-- 013_api_keys.sql
-- Purpose: API key authentication for tenant-scoped access.

CREATE TABLE IF NOT EXISTS tenant_api_keys (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    uuid NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  name         text,
  key_prefix   text,
  key_hash     text NOT NULL UNIQUE,
  roles        text[] NOT NULL DEFAULT '{}',
  created_at   timestamptz NOT NULL DEFAULT now(),
  expires_at   timestamptz,
  revoked_at   timestamptz,
  last_used_at timestamptz
);

CREATE INDEX IF NOT EXISTS tenant_api_keys_tenant_idx
ON tenant_api_keys (tenant_id);

CREATE INDEX IF NOT EXISTS tenant_api_keys_active_idx
ON tenant_api_keys (tenant_id)
WHERE revoked_at IS NULL;
