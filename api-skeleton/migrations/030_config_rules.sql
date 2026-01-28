-- 030_config_rules.sql
-- Purpose: versioned tenant config + rules with "current pointers".
-- Config/rules are stored as jsonb. ETags are computed from config::text
-- using sha256 to support optimistic concurrency.

CREATE TABLE IF NOT EXISTS tenant_config_versions (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id      uuid NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  config         jsonb NOT NULL,
  etag           text GENERATED ALWAYS AS (
                   encode(digest(config::text, 'sha256'), 'hex')
                 ) STORED,
  created_at     timestamptz NOT NULL DEFAULT now(),
  created_by_sub text NOT NULL,
  comment        text,
  UNIQUE (tenant_id, etag)
);

DROP TRIGGER IF EXISTS tenant_config_versions_immutable_ud
ON tenant_config_versions;
CREATE TRIGGER tenant_config_versions_immutable_ud
BEFORE UPDATE OR DELETE ON tenant_config_versions
FOR EACH ROW EXECUTE FUNCTION prevent_update_delete();

CREATE INDEX IF NOT EXISTS tenant_config_versions_tenant_created_idx
ON tenant_config_versions (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS tenant_config_versions_config_gin
ON tenant_config_versions USING gin (config);

CREATE TABLE IF NOT EXISTS tenant_configs_current (
  tenant_id         uuid PRIMARY KEY REFERENCES tenants(id) ON DELETE RESTRICT,
  config_version_id uuid NOT NULL REFERENCES tenant_config_versions(id)
                   ON DELETE RESTRICT,
  updated_at        timestamptz NOT NULL DEFAULT now(),
  updated_by_sub    text NOT NULL
);

DROP TRIGGER IF EXISTS tenant_configs_current_set_updated_at
ON tenant_configs_current;
CREATE TRIGGER tenant_configs_current_set_updated_at
BEFORE UPDATE ON tenant_configs_current
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE TABLE IF NOT EXISTS tenant_rule_versions (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id      uuid NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  rules          jsonb NOT NULL,
  etag           text GENERATED ALWAYS AS (
                   encode(digest(rules::text, 'sha256'), 'hex')
                 ) STORED,
  created_at     timestamptz NOT NULL DEFAULT now(),
  created_by_sub text NOT NULL,
  comment        text,
  UNIQUE (tenant_id, etag)
);

DROP TRIGGER IF EXISTS tenant_rule_versions_immutable_ud
ON tenant_rule_versions;
CREATE TRIGGER tenant_rule_versions_immutable_ud
BEFORE UPDATE OR DELETE ON tenant_rule_versions
FOR EACH ROW EXECUTE FUNCTION prevent_update_delete();

CREATE INDEX IF NOT EXISTS tenant_rule_versions_tenant_created_idx
ON tenant_rule_versions (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS tenant_rule_versions_rules_gin
ON tenant_rule_versions USING gin (rules);

CREATE TABLE IF NOT EXISTS tenant_rules_current (
  tenant_id        uuid PRIMARY KEY REFERENCES tenants(id) ON DELETE RESTRICT,
  rules_version_id uuid NOT NULL REFERENCES tenant_rule_versions(id)
                  ON DELETE RESTRICT,
  updated_at       timestamptz NOT NULL DEFAULT now(),
  updated_by_sub   text NOT NULL
);

DROP TRIGGER IF EXISTS tenant_rules_current_set_updated_at
ON tenant_rules_current;
CREATE TRIGGER tenant_rules_current_set_updated_at
BEFORE UPDATE ON tenant_rules_current
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
