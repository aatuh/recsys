---
diataxis: reference
tags:
  - reference
  - api
  - versioning
  - compatibility
---
# API compatibility policy
This page is the canonical reference for API compatibility policy.


## Who this is for

- Lead developers and platform engineers integrating the API long-term.
- Buyers who need procurement-grade clarity about “what can change”.

## What you will get

- A clear stability contract for request/response schemas
- What changes are safe vs breaking
- How versioning works in this suite

## Example (request shape tolerated across minors)

```http
POST /v1/recommend HTTP/1.1
Content-Type: application/json

{
  "tenant_id": "demo",
  "surface": "home",
  "k": 10
}
```

## Versioning model

RecSys uses:

- **SemVer** for releases (suite version)
- A versioned HTTP API namespace (`/v1/...`) for serving and admin endpoints

See also:

- Suite docs versioning: [Docs versioning](../../project/docs-versioning.md)

## Compatibility guarantees

Within a stable major version:

- **Minor versions may add fields** to JSON objects (responses and requests) without breaking clients.
- **New endpoints may be added** without breaking existing clients.
- **Behavioral changes** that affect ranking output should be gated through evaluation and documented in “What’s new”.

Breaking changes (require a new major version):

- Removing or renaming fields that existing clients rely on
- Changing types or semantics of existing fields
- Removing or changing meaning of error codes

## Client guidance (how to integrate safely)

- Treat unknown JSON fields as ignorable.
- Pin your dependency versions and read “What’s new” on upgrades.
- For experiments, ensure deterministic assignment inputs are stable across platforms.

Docs:

- What’s new: [What’s new](../../whats-new/index.md)
- Experimentation model: [Experimentation model (A/B, interleaving, OPE)](../../explanation/experimentation-model.md)

## Read next

- API reference: [API Reference](api-reference.md)
- Errors: [Error handling & troubleshooting API calls](errors.md)
- Admin API: [Admin API + local bootstrap (recsys-service)](admin.md)
