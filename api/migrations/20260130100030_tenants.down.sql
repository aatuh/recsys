-- 010_tenants.sql (down)
DROP TRIGGER IF EXISTS tenants_set_updated_at ON tenants;
DROP TABLE IF EXISTS tenants;
