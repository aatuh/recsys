# Tutorial: local end-to-end (pipelines -> service -> eval)

Goal: run a tiny dataset through pipelines, serve recommendations, and generate
an evaluation report.

Prereqs:
- Docker + docker compose
- curl
- POSIX shell

## 1. Start local dependencies

Bring up Postgres and recsys-service. Ensure migrations are applied.

## 2. Load demo tenant + config + rules

Create:
- tenant: demo
- config and rules (versioned, current pointer updated)

See: `reference/api/examples/admin-config.http`
See also: `reference/api/admin.md` (auth + bootstrap details, tenant insert SQL)

## 3. Load a tiny dataset

Use `tutorials/datasets/tiny/`:
- catalog.csv
- exposures.jsonl
- interactions.jsonl

In production, pipelines ingest raw events. For this tutorial you can import the
files into your canonical store.

## 4. Run pipelines to publish artifacts

Run the minimum jobs to produce non-empty recs (e.g. popularity + manifest).

See: `how-to/operate-pipelines.md`

## 5. Call the API

- POST /v1/recommend
- POST /v1/similar (optional)
- POST /v1/recommend/validate (lint requests)

See: `reference/api/examples/`

## 6. Run evaluation

Run offline regression and (optionally) experiment analysis.

See: `how-to/run-eval-and-ship.md`
