# How-to: operate recsys-pipelines

Daily:
- ingest exposures + outcomes
- validate/canonicalize
- build artifacts (start with popularity)
- publish atomically (update manifest pointer)
- monitor freshness, volume anomalies, output sizes

Backfills:
- compute for explicit time windows
- publish new manifest version
- keep prior manifest for rollback

Rollback:
- pointer swap to last good manifest
- invalidate service caches

DB-only mode:
- write signals into Postgres tables instead of publishing artifacts
- useful for local MVPs and popularity-only pilots
- seed examples: `reference/database/db-only-seeding.md`
