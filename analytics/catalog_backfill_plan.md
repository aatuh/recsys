# Catalog Backfill & Freshness Plan

Purpose: ensure all existing items receive the newly required metadata + embeddings and stay fresh going forward.

## 1. Backfill Overview
1. Snapshot current catalog from master data source (Shop DB or product feed) including `brand`, `category`, `description`, `image_url`, pricing, availability.
2. Materialise snapshot into staging table `analytics.catalog_snapshot_YYYYMMDD`.
3. Run enrichment scripts:
   - **Normalize** brand/category to canonical values.
   - **Derive** `category_path` (split taxonomy, ensure depth ≤ 4).
   - **Set** `metadata_version` = snapshot date or upstream version hash.
4. Bulk update `items` table via batched `items:upsert` API calls (or direct SQL) to populate new columns.
5. Kick off embedding generation job (`analytics/embedding_pipeline.md`) for all items with missing embeddings.

## 2. Freshness Workflow
- **Daily diff job** `catalog_metadata_refresh`:
  1. Fetch items changed in upstream product feed (by `metadata_version`/`updated_at`).
  2. Upsert metadata via ingestion service.
  3. Schedule embedding regeneration for changed items.
- **Real-time hook** (optional): when Shop CMS publishes item updates, emit webhook that POSTs to `/v1/items:upsert`.
- **Monitoring**: expose Prometheus metrics
  - `catalog_metadata_missing_total` (items missing brand/category/description).
  - `catalog_embeddings_stale_total` (embeddings older than threshold).
  - Alerts when metrics exceed thresholds or backfill job fails.

## 3. Rollout Checklist
- [ ] Create staging schema + dbt models for catalog snapshot.
- [ ] Write Go/dbt script to batch entries through ingestion API (respect rate limits).
- [ ] Configure Airflow schedule for daily refresh + embedding regeneration.
- [ ] Add Grafana panels for metadata coverage (baseline 100% brand/category, ≥ 90% description).
- [ ] Document rollback: clear new fields by re-running snapshot from previous stable version.

## 4. Ownership
- **Data Engineering**: snapshot + normalization jobs.
- **Ranking Platform**: ingestion + monitoring.
- **ML/Feature Store**: embedding job + vector storage maintenance.
