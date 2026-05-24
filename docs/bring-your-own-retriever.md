# Bring Your Own Retriever

RecSys is batch/offline-first. It can consume embedding and candidate signals, but it is not a built-in vector database
or feature store. For embedding-heavy deployments, use an external retriever to produce candidates, then let RecSys apply
deterministic ranking, rules, diversity, explanations, exposure logging, and evaluation.

## Contract

The retriever owns:

- embedding generation and refresh,
- vector/ANN index build and serving,
- feature-store or warehouse joins,
- retrieval latency and fallback behavior,
- retriever-specific monitoring.

RecSys owns:

- request validation and tenant-aware serving,
- deterministic ranking and merchandising rules,
- artifact rollout and rollback,
- exposure metadata,
- offline evaluation inputs and decision gates.

## Request Shape

Send retrieved candidates through the existing recommendation request:

```json
{
  "surface": "home",
  "k": 10,
  "user": {"anonymous_id": "anon-123", "session_id": "sess-456"},
  "candidates": {
    "include_ids": ["sku-1", "sku-2", "sku-3"],
    "exclude_ids": ["sku-hidden"]
  },
  "options": {"include_reasons": true}
}
```

Keep candidate IDs tenant-safe and catalog-valid. Do not send raw embeddings, email addresses, names, phone numbers, or
other direct PII in request payloads.

## Failure Behavior

If the retriever is unavailable, use one of these explicit fallbacks:

- call RecSys without `candidates.include_ids` and rely on artifact-backed popularity/co-occurrence,
- use cached retriever output with a short TTL,
- return merchandising fallback content in the client.

Log which fallback path was used so evaluation can exclude or slice degraded traffic.

## Evaluation

Keep retriever version, index version, RecSys algorithm version, rules version, and manifest version in decision records.
Do not ship a retriever change on KPI movement alone; first verify schema validity, join integrity, latency, errors,
empty recommendation rate, and rollback readiness.
