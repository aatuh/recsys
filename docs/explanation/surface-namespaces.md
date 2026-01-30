# Surface namespaces

The service treats **surface** as a first-class scoping key. In practice this
means the online API uses `surface` as the **namespace** when reading signals
(e.g., popularity, tags, co-vis). This keeps signals isolated per surface by
default and prevents "home" data from leaking into "pdp" results.

## Recommended strategy

- Use a small, stable list of surface names (e.g., `home`, `pdp`, `cart`).
- In pipelines, emit artifacts/signals with `surface` matching the API surface.
- In DB-only mode, write signals under `namespace = <surface>`.

## Default namespace fallback

The service currently **falls back to `default`** for popularity/tags when a
surface namespace is missing. This is helpful for local development but can be
surprising in production.

Best practice:
- Seed per-surface data for production.
- Use `default` only for shared or global signals.

Note on similar-items:
- `/v1/similar` reads co-visit signals scoped by **surface/namespace**.
- If you only seed `home`, requests for `surface=pdp` will return empty results.

## Example

- API request: `surface=home`
- Signals stored under: `namespace=home`
- Optional fallback: `namespace=default` if `home` has no rows
