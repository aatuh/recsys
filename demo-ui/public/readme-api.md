# README

An HTTP API for ingesting activity and asking for recommendations.

## What It Does

- **Ingest** items, users, events using **opaque IDs** (you keep the mapping).
- **Recommend** top‑K items using **time‑decayed popularity** (v1).
- **Similar items** via **co‑visitation**.
- **Per‑tenant config** for event types, weights, and optional per‑type
  half‑life.

## Endpoints (core)

Open **Swagger** at **`/docs`** to inspect schemas and try requests.

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
