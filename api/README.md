# README

An HTTP API for ingesting activity and asking for recommendations.

## What It Does

- **Ingest** items, users, events using **opaque IDs** (you keep the mapping).
- **Recommend** top‑K items using **time‑decayed popularity** (v1).
- **Similar items** via **co‑visitation**.
- **Per‑tenant config** for event types, weights, and optional per‑type
  half‑life.
- **Audit trail** that captures each recommendation decision for compliance and
  debugging.

## Endpoints (core)

Open **Swagger** at **`/docs`** to inspect schemas and try requests.

### Audit trail

- `GET /v1/audit/decisions` — list recent decisions with filters for namespace,
  time range, user hash, or request id.
- `GET /v1/audit/decisions/{decision_id}` — fetch the full stored trace for a
  single decision.

## How It Works

- **Popularity**: sum of (`event_weight` × half‑life decay × optional `value`)
  per item.
  - Weights & (optional) per‑type half‑life come from tenant overrides, falling
    back to global defaults.
- **Co‑visitation**: count recent co‑occurring items by the same users.

## Development

- See the **Makefile** in the repo root for dev/test/migration commands
  (e.g., `make dev`, `make test`).  
- Hot reload in dev; Swagger is generated from annotations.

## Deploying to Railway

- Create new project.
- Paste env variables.
- Set connected branch as `production`.
- Root directory `/api/`.
- Generate custom domain.
