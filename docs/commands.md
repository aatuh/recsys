# Offline Command Guide

Use this reference when you need to run internal Go binaries that operate outside the public API surface. These commands live under `api/cmd/` and must be run from the `api/` directory (for example, `cd api && go run ./cmd/catalog_backfill ...`). They touch internal tables directly, so treat them as privileged maintenance tools.

> **Who should read this?** Engineers and operators who manage RecSys namespaces, seed data, or validate ranking configs locally.
>
> **Where this fits:** Control-plane maintenance & tuning. Pair this guide with [`docs/tuning_playbook.md`](tuning_playbook.md) for end-to-end workflows and [`docs/analysis_scripts_reference.md`](analysis_scripts_reference.md) for the Python automation catalog.

### TL;DR

- `catalog_backfill` re-derives catalog metadata/embeddings directly in the database—use it after ingestion changes or schema tweaks.
- `collab_factors` regenerates item/user factor vectors from historical events to keep collaborative filtering signals fresh.
- `blend_eval` runs the scoring harness offline to benchmark blend weights and Maximal Marginal Relevance (MMR, a diversity-aware re-ranking method) knobs without hitting the HTTP API.

---

## 1. Catalog backfill (`api/cmd/catalog_backfill`)

**Purpose.** Recompute brand/category/category-path metadata, normalize descriptions/image URLs, bump metadata versions, and optionally synthesize deterministic embeddings for catalog rows. It uses `store.CatalogItems` plus `internal/catalog.BuildUpsert`, so it can see raw props that never flow through the public ingestion APIs.

**Run it when**
- You ingested raw rows with missing or stale derived fields (brand, category paths, metadata version).
- You changed derivation logic in `internal/catalog` or updated embedding behavior and need to reapply it.
- You want a scheduled sweep (`--mode refresh --since 24h`) that keeps catalog hygiene without reprocessing everything.

**Command**

```bash
cd api
go run ./cmd/catalog_backfill \
  --namespace retail_demo \
  --mode refresh \
  --since 24h \
  --batch 500
```

**Key flags**
- `--mode backfill` scans only rows missing derived data; `--mode refresh` reprocesses whatever changed since `--since`.
- `--since` accepts durations (`24h`, `7d`) or RFC3339 timestamps for precise windows.
- `--dry-run` logs the upserts (item ID + payload) so you can verify changes before writing.

---

## 2. Collaborative factors (`api/cmd/collab_factors`)

**Purpose.** Generate dense vectors for items and users directly from the `items` and `events` tables. Items reuse existing embeddings or build deterministic fallbacks; users average the embeddings of items they interacted with. The command persists the result via `store.UpsertItemFactors` and `store.UpsertUserFactors`, which have no public API equivalent.

**Run it when**
- You seeded or imported a large batch of events and want collaborative filtering signals to reflect the new activity.
- You run periodic refreshes (for example, nightly `--since 7d`) instead of recomputing factors from scratch.
- You plan to evaluate blend weights that rely on collaborative factors and need up-to-date vectors.

**Command**

```bash
cd api
go run ./cmd/collab_factors \
  --namespace retail_demo \
  --since 30d
```

**Helpful flags**
- `--since` limits the event range for incremental runs; omit it to rebuild from all historical events.
- `--dry-run` surfaces totals (items/users processed/upserted) without writing, which is useful to confirm embedding lengths or namespace coverage.

---

## 3. Blend evaluation harness (`api/cmd/blend_eval`)

**Purpose.** Exercise the internal blend scorer offline. You can compare multiple candidate configs (different alpha/beta/gamma weights, MMR settings, profile boosts) against historical ground truth without making HTTP calls. The harness prints hit rate, Mean Reciprocal Rank (MRR, “how early do good items appear?”), average rank, coverage, and average list length per candidate.

**Run it when**
- You are about to change default blend weights or publish a new namespace profile.
- You want a quick regression check without seeding data + running the full analysis simulation pipeline.
- You are experimenting inside a sandbox where the public API or quality scripts are unavailable.

**Command**

```bash
cd api
go run ./cmd/blend_eval \
  --namespace retail_demo \
  --limit 400 \
  --min-events 10 \
  --configs analysis/fixtures/templates/marketplace.json
```

**Helpful flags**
- `--configs` points to a YAML file with a `configs` list of blend candidates; omit it to use baked-in defaults.
- `--lookback 720h` widens the event window for sparse namespaces.
- `--k` controls the list depth evaluated for hit rate and coverage.

---

## Where to go next

- [`README.md`](../README.md) – top-level commands (`make dev`, `make test`) and repo map for running maintenance workflows.
- `analysis/scripts/run_quality_eval.py` – HTTP-based evaluation suite when you need evidence bundles or guardrail checks tied to the API.
