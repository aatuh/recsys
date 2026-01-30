-- 030_config_rules.sql (down)
DROP TRIGGER IF EXISTS tenant_rules_current_set_updated_at ON tenant_rules_current;
DROP TABLE IF EXISTS tenant_rules_current;

DROP TRIGGER IF EXISTS tenant_rule_versions_immutable_ud ON tenant_rule_versions;
DROP TABLE IF EXISTS tenant_rule_versions;

DROP TRIGGER IF EXISTS tenant_configs_current_set_updated_at ON tenant_configs_current;
DROP TABLE IF EXISTS tenant_configs_current;

DROP TRIGGER IF EXISTS tenant_config_versions_immutable_ud ON tenant_config_versions;
DROP TABLE IF EXISTS tenant_config_versions;
