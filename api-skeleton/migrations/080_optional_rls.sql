-- 080_optional_rls.sql
-- Purpose: optional row-level security (RLS) "seatbelt".
-- Enable only if you plan to set app.tenant_id (or similar) per connection
-- and understand RLS bypass rules for owners/superusers.

-- Example:
-- ALTER TABLE tenant_config_versions ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE tenant_configs_current ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE tenant_rule_versions ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE tenant_rules_current ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE cache_invalidation_events ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE exposure_events ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE interaction_events ENABLE ROW LEVEL SECURITY;

-- Example policy pattern (repeat per table):
-- CREATE POLICY tenant_isolation_exposure
-- ON exposure_events
-- USING (tenant_id::text = current_setting('app.tenant_id', true));

-- Optional: force RLS even for table owners:
-- ALTER TABLE exposure_events FORCE ROW LEVEL SECURITY;
