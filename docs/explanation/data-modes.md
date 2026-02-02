# Data modes: DB-only vs object store

The service supports **DB-only mode** and **artifact/manifest mode**. DB-only
is the default; artifact mode is opt-in via config.

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

Enable artifact mode:

- `RECSYS_ARTIFACT_MODE_ENABLED=true`
- `RECSYS_ARTIFACT_MANIFEST_TEMPLATE` (e.g. `s3://recsys/registry/current/{tenant}/{surface}/manifest.json`

  or `file:///data/registry/current/{tenant}/{surface}/manifest.json`)

Notes:

- `{tenant}` uses the incoming tenant id (header/JWT) when available.
- `{surface}` maps to the request surface (namespace).
- Tags and constraints still read from Postgres (`item_tags`), even in artifact mode.

## Recommendation

- Use **DB-only** for MVP and local testing (default today).
- Use **object store + manifest** for production-scale artifacts once the

  pipelines are producing artifacts and the service is configured to read them.

## Which mode is active?

The service runs in **DB-only mode by default**. When
`RECSYS_ARTIFACT_MODE_ENABLED=true`, the service reads popularity/co-visitation
from the artifact manifest and uses Postgres for tag metadata.
