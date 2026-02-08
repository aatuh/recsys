---
diataxis: reference
tags:
  - reference
  - api
  - developer
  - recsys-service
---
# Recommend request fields

This page is a **field-level reference** for `POST /v1/recommend` request payloads.

## Who this is for

- Backend/frontend engineers integrating `POST /v1/recommend`.
- Platform engineers defining stable request contracts across clients.

## What you will get

- Required fields and optional request sections.
- Minimal schema shape and per-field semantics.
- Links to validation and related ranking/tenancy references.

## Location

- OpenAPI spec: `docs/reference/api/openapi.yaml`
- Swagger UI: [OpenAPI / Swagger UI](api-reference.md)

## Quick schema shape

```json
{
  "tenant_id": "...",
  "surface": "...",
  "k": 10,
  "user": { "user_id": "..." },
  "anchors": { "item_ids": ["..."] },
  "candidates": { "include_ids": [], "exclude_ids": [] },
  "weights": { "pop": 1.0, "cooc": 1.0, "emb": 1.0 },
  "constraints": { "required_tags": [], "forbidden_tags": [], "max_per_tag": {} },
  "options": { "include_reasons": true, "explain": "summary" },
  "context": {}
}
```

Only a subset is required. Use `/v1/recommend/validate` to see the normalized request.

---

## Required fields

- `tenant_id` (string)
  - Tenancy boundary used for config + data isolation.
- `surface` (string)
  - The recommendation surface you are requesting (home, pdp_similar, etc.).
- `k` (integer)
  - Number of items requested.

See tenancy semantics: [Auth and tenancy](../auth-and-tenancy.md)

## `user`

- `user.user_id` (string)
  - Stable pseudonymous identifier.

## `anchors`

Use anchors when a request has a natural seed item (similar-items, co-visitation, session-based recs).

- `anchors.item_ids` (array of strings)
  - Explicit seed items.

## `candidates`

Candidate list controls are **filters** and do not replace normal candidate generation.

- `candidates.include_ids` (array of strings)
  - Final allow-list filter.
  - If it removes everything you will see `CANDIDATES_INCLUDE_EMPTY`.
- `candidates.exclude_ids` (array of strings)
  - Strict exclusion list.

## `weights`

Weights control signal blending per request.

- `weights.pop` (number)
  - Popularity contribution.
- `weights.cooc` (number)
  - Co-visitation contribution.
- `weights.emb` (number)
  - Similarity contribution (non-popularity, non-co-visitation bucket).

If omitted, defaults come from config.

Config defaults live here:

- [Ranking & constraints reference](../../recsys-algo/ranking-reference.md)

## `constraints`

Constraints apply after ranking (when tag data is available).

- `constraints.required_tags` (array of strings)
  - Require at least one matching tag.
- `constraints.forbidden_tags` (array of strings)
  - Filter out items with these tags.
- `constraints.max_per_tag` (object)
  - Enforce diversity caps, e.g. `{ "category:shoes": 2 }`.

## `options`

Options control explainability and output metadata.

- `options.include_reasons` (boolean)
  - Adds `reasons[]` per item.
- `options.explain` (string)
  - `none|summary|full`.

## `context`

`context` carries downstream slicing keys.

Recommended keys:

- `tenant_id`, `surface`, `segment`
- experiment identifiers (if you run A/B tests)

## Validation endpoint

Before you integrate, use:

- `POST /v1/recommend/validate`

It returns a normalized request and early warnings.

## Related

- Explanation: [Candidate generation vs ranking](../../explanation/candidate-vs-ranking.md)
- How-to: [Integrate recsys-service](../../how-to/integrate-recsys-service.md)
