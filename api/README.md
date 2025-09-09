# README

An HTTP API for ingesting activity and asking for recommendations.

## What it does

- **Ingest** items, users, events using **opaque IDs** (you keep the mapping).
- **Recommend** top‑K items using **time‑decayed popularity** (v1).
- **Similar items** via **co‑visitation**.
- **Per‑tenant config** for event types, weights, and optional per‑type
  half‑life.

## Endpoints (core)

- `POST /v1/items:upsert` - batch create/update items
- `POST /v1/users:upsert` - batch create/update users
- `POST /v1/events:batch` - batch user→item events
- `POST /v1/recommendations` - top‑K by popularity (v1)
- `GET  /v1/items/{item_id}/similar?namespace=…&k=…` - co‑vis neighbors
- `POST /v1/event-types:upsert` / `GET /v1/event-types` - tenant overrides

Open **Swagger** at **`/docs`** to inspect schemas and try requests.

## How it works (ranking v1)

- **Popularity**: sum of (`event_weight` × half‑life decay × optional `value`)
  per item.
  - Weights & (optional) per‑type half‑life come from tenant overrides, falling
    back to global defaults.
- **Co‑visitation**: count recent co‑occurring items by the same users.

## Development

- See the **Makefile** in the repo root for dev/test/migration commands
  (e.g., `make dev`, `make test`).  
- Hot reload in dev; Swagger is generated from annotations.
