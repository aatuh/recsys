-- 002_types.sql
-- Purpose: domain enums for strictness and consistency.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tenant_status') THEN
    CREATE TYPE tenant_status AS ENUM ('active', 'suspended', 'deleted');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'api_client_status') THEN
    CREATE TYPE api_client_status AS ENUM ('active', 'revoked');
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_type WHERE typname = 'cache_invalidation_status'
  ) THEN
    CREATE TYPE cache_invalidation_status AS ENUM (
      'requested', 'applied', 'failed'
    );
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'interaction_type') THEN
    CREATE TYPE interaction_type AS ENUM (
      'impression', 'click', 'add_to_cart', 'purchase', 'view'
    );
  END IF;
END $$;
