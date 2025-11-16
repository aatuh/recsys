# Catalog Metadata Inventory

Goal: document the current product attributes powering retrievers/rankers, highlight gaps for content-based models, and specify what needs to be ingested before heading into the embedding/backfill work.

## 1. Current State

- **`item_id`** — Source: shop ingestion (`items:upsert`). Storage: `items` table. Consumers: all retrievers/rankers.
- **`available`** — Source: shop ingestion. Storage: `items.available`. Consumers: eligibility filters.
- **`price`** — Source: shop ingestion (`price` optional). Storage: `items.price` (numeric). Consumers: popularity constraints, merchandising overrides.
- **`tags`** — Source: ingestion `tags` array. Storage: `items.tags` (text[]) plus `items_tags` view. Consumers: content retriever, MMR caps, personalization profiles.
- **`props`** — Source: optional JSON blob. Storage: `items.props` (jsonb). Consumers: none yet.
- **`created_at`** — Source: derived. Storage: `items.created_at`. Consumers: cold-start freshness logic.
- **Embeddings** — Not available yet (no storage/consumers).
- **Rich metadata (brand/category taxonomy, text, images)** — Partial via tags in `items.tags` with prefixes. Consumers: diversity caps, explainability tooling.

## 2. Gaps and Requirements

1. **Textual Descriptions** – Needed for embedding generation (OpenAI / custom). Should be stored in a new column (`items.description`) or props key.
2. **Category Hierarchy** – Current tags rely on `brand:`/`category:` prefixes. Need formal columns (`brand`, `category`, `category_path`) to ease analytics and similarity joins.
3. **Image URLs / Feature Vectors** – Required for visual embeddings. Store URLs in `props.media` and plan for vector store integration.
4. **Inventory Signals** – Stock levels or click-through on detail pages; currently absent. Optional for weighting future retrievers.
5. **Language / Locale** – For multi-language support; currently not captured.

## 3. Proposed Data Additions (Pre-embedding rollout)

- **`description`** — Type: text. Source: shop CMS export. Notes: feed textual embedding pipeline (sanitize input).
- **`brand`** — Type: text. Source: normalized tags/product feed. Notes: replace `brand:` prefix usage.
- **`category`** — Type: text. Source: normalized canonical taxonomy. Notes: enables analytics + retrieval filters.
- **`category_path`** — Type: text[]. Source: derived. Notes: enables hierarchical similarity.
- **`image_url`** — Type: text. Source: product media service. Notes: input for visual embedding pipeline.
- **`props.metadata_version`** — Type: text. Source: ingestion job. Notes: track metadata freshness.

## 4. Data Pipeline Owners

- **Product ingestion**: Shop backend to emit enriched payloads; extend `/items:upsert` contract.
- **Backfill**: run one-time script (dbt or Go) to hydrate new columns from existing tags + upstream feed.
- **Validation**: add schema checks in `api/test/store` to ensure new columns populated for sample fixtures.

## 5. Next Steps

1. Update swagger/specs to include optional `brand`, `category`, `description`, `image_url` fields.
2. Coordinate with data engineering to expose product feed with the required fields.
3. Define embedding generation pipeline (tooling, batch schedule).
4. Update runbooks (diversity/personalisation dashboards) to incorporate new metadata for monitoring.
