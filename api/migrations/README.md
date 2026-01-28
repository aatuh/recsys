# recsys-svc Postgres schema migrations

This package contains baseline Postgres migrations for recsys-svc.

## Apply order
Run the SQL files in lexical order from `migrations/`.

Example with psql:

  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/001_extensions.sql
  ...

## Notes
- This schema supports:
  - Tenants
  - Versioned tenant config/rules (append-only) + "current" pointers + ETags
  - Admin audit log (append-only)
  - Cache invalidation tracking
  - Exposure and interaction event logs (partitioned by time)
  - Optional API client keys (if you support them)

- Partition maintenance:
  - The event tables include DEFAULT partitions to avoid insert failures.
  - Use `migrations/070_partition_helpers.sql` to create monthly partitions
    ahead of time.

- Optional RLS:
  - `migrations/080_optional_rls.sql` includes commented examples for enabling
    row-level security as an extra "seatbelt" for tenant isolation.

## Compatibility
- Assumes Postgres 14+ (works on newer versions too).
