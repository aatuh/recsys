# Migrations (safe upgrade)

## Policy

- Migrations are **versioned, append-only**. Never edit a migration after it has
  shipped; create a new one instead.
- `down` migrations are **disabled by default** in production. Use only in
  controlled rollback scenarios.

## Preflight checks

Before upgrading from N‑1 to N, run:

```
cd recsys/api
go run ./cmd/migrate preflight
```

Docker/compose shortcut:

```
cd recsys/api
make migrate-preflight
```

This verifies:
- DB connectivity
- No failed migrations in `schema_migrations`
- Checksums match the local migration files

## Upgrade steps (N‑1 → N)

1. Take a DB snapshot/backup.
2. Run preflight checks.
3. Apply migrations:

```
go run ./cmd/migrate up
```

4. Verify status:

```
go run ./cmd/migrate status
```

5. Roll forward with application deploy.

## Rollback story

If a migration introduces issues:

1. Roll back the application to N‑1.
2. Use the config/rules rollback runbook:
   `docs/operations/runbooks/rollback-config-rules.md`
3. Only if absolutely required, apply a **controlled** `down` migration:

```
go run ./cmd/migrate --allow-down down
```

## Recommended order (fresh install)

1. extensions
2. tenants
3. config/rules version tables + current pointers
4. audit log
5. exposure_events (partitioned)
