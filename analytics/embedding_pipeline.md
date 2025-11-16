# Product Embedding Pipeline Plan

Goal: generate textual and (optionally) visual embeddings for catalog items, store them in a feature service, and expose them to the content-based retriever.

## 1. Source Data

- **Required fields** (from the catalog inventory plan): `item_id`, `description`, `brand`, `category`, `image_url`.
- Pull from product feed / ingestion service into a staging table (`analytics.catalog_features_raw`).

## 2. Embedding Generation

- **Text** — Model: OpenAI `text-embedding-3-large` (or local alternative). Schedule: nightly batch (dbt + Python). Output: 3072-d vector.
- **Image (optional)** — Model: CLIP / ViT. Schedule: weekly. Output: 512-d vector.

**Process**
1. Extract items needing refresh (new or `metadata_version` changed).
2. Chunk text inputs (< 8k tokens) with prompt: `"Product: {brand} {category}\nDescription: {description}"`.
3. Call embedding API with exponential backoff; store raw vectors in object storage (parquet) and push average vector to Postgres via new table `item_embeddings`.
4. Maintain `embedding_version`, `generated_at`. Allow multiple versions for rollback/testing.

## 3. Storage Schema (`item_embeddings`)

```sql
CREATE TABLE item_embeddings (
  org_id UUID NOT NULL,
  item_id TEXT NOT NULL,
  namespace TEXT NOT NULL DEFAULT 'default',
  version TEXT NOT NULL,
  modality TEXT NOT NULL CHECK (modality IN ('text','image')),
  dims INT NOT NULL,
  vector VECTOR(3072),
  generated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (org_id, item_id, namespace, modality, version)
);
```

- Consider using pgvector for efficient cosine similarity.

## 4. Retriever Integration

1. Extend `internal/store/content.go` to query `item_embeddings` when tags insufficient (`JOIN item_embeddings` by namespace + version).
2. Update `CandidateData` to include embedding-derived similarity scores (already partly handled in RT-1C).
3. Add config for selecting embedding version per namespace (`BLEND_EMBEDDING_VERSION`).

## 5. Pipeline Orchestration

- Use Airflow/dbt job: `generate_product_embeddings`.
- Steps: extract → transform (prompt formatting) → call API → load results.
- Monitoring: log success counts, failure rate, average latency. Alert if failure >5%.

## 6. Cost & Performance Considerations

- Batch within API rate limits (parallelism <= 10). Cache previous outputs (`metadata_hash`).
- Store compressed vectors (pgvector uses ~12 bytes/dim); plan disk accordingly.

## 7. Rollout Plan

1. Backfill 5% of catalog (A/B) to validate retriever lift.
2. Monitor blend experiments for CTR/personalisation impact.
3. Ramp to full catalog and enable in production blend weights.

## 8. Owners

- Data Platform: embedding batch job.
- ML Engineer: model selection/hyperparams.
- Ranking Team: retriever integration and runtime config.
