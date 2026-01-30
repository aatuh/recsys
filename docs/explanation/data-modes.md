# Data modes: DB-only vs object store

The service supports **DB-only mode** today and is designed to later support
**artifact/manifest mode** once the online reader is wired.

## DB-only mode (current, recommended for MVP)

Signals are stored directly in Postgres tables and read by the service:

- `item_tags`
- `item_popularity_daily`
- `item_covisit_daily` (if enabled)

Popularity uses a decayed sum over `item_popularity_daily` with the configured
half-life, so **newer days dominate** when you seed both recent and older rows.

This is ideal for local development and popularity-only pilots.

## Artifact/manifest mode (pipelines + object store)

Pipelines can publish artifacts (popularity, co-vis, embeddings) to object
storage and update a manifest pointer. This enables atomic updates and easy
rollback, but the **service must be configured to read artifacts**.

## Recommendation

- Use **DB-only** for MVP and local testing.
- Use **object store + manifest** for production-scale artifacts once the
  service reader is available.
