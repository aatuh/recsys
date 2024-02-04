# Migrations (safe upgrade)

## Policy

- Migrations are **versioned, append-only**. Never edit a migration after it has

  shipped; create a new one instead.

- `down` migrations are **disabled by default** in production. Use only in

  controlled rollback scenarios.

## Preflight checks

Before upgrading from N‑1 to N, run:

```bash
cd api
go run ./cmd/migrate preflight
```

Docker/compose shortcut:

```bash
cd api
make migrate-preflight
```

This verifies:

- DB connectivity
- No failed migrations in `schema_migrations`
- Checksums match the local migration files

## Upgrade steps (N‑1 → N)

1. Take a DB snapshot/backup.
1. Run preflight checks.
1. Apply migrations:

```bash
go run ./cmd/migrate up
```

1. Verify status:

```bash
go run ./cmd/migrate status
```

1. Roll forward with application deploy.

## Rollback story

If a migration introduces issues:

1. Roll back the application to N‑1.
2. Use the config/rules rollback runbook:

   `docs/operations/runbooks/rollback-config-rules.md`

3. Only if absolutely required, apply a **controlled** `down` migration:

```bash
go run ./cmd/migrate --allow-down down
```

## Recommended order (fresh install)

1. extensions
1. tenants
1. config/rules version tables + current pointers
1. audit log
1. exposure_events (partitioned)
