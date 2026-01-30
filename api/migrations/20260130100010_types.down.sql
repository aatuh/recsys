-- 002_types.sql (down)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'interaction_type') THEN
    DROP TYPE interaction_type;
  END IF;
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'cache_invalidation_status') THEN
    DROP TYPE cache_invalidation_status;
  END IF;
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'api_client_status') THEN
    DROP TYPE api_client_status;
  END IF;
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tenant_status') THEN
    DROP TYPE tenant_status;
  END IF;
END $$;
