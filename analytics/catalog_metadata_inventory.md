# Catalog Metadata Inventory (RT-5A)

Goal: document the current product attributes powering retrievers/rankers, highlight gaps for content-based models, and specify what needs to be ingested before heading into RT-5B-5D.

## 1. Current State

| Attribute | Source | Storage | Consumers |
|-----------|--------|---------|-----------|
| `item_id` | Shop ingestion (`items:upsert`) | `items` table | All retrievers/rankers |
| `available` | Shop ingestion | `items.available` | Eligibility filters |
| `price` | Shop ingestion (`price` optional) | `items.price` (numeric) | Popularity constraints, merchandising overrides |
| `tags` | Array supplied via ingestion (`tags`) | `items.tags` (text[]) + `items_tags` view | Content retriever, MMR caps, profile/
| `props` | JSON blob (optional) | `items.props` (jsonb) | Not actively used yet |
| `created_at` | Derived | `items.created_at` | Cold-start freshness |
| Embeddings | **Not available** | — | — |
| Rich metadata (brand/category taxonomy, text, images) | Partial via tags | `items.tags` (prefixed values) | Diversity caps, explainability |

## 2. Gaps and Requirements

1. **Textual Descriptions** – Needed for embedding generation (OpenAI / custom). Should be stored in a new column (`items.description`) or props key.
2. **Category Hierarchy** – Current tags rely on `brand:`/`category:` prefixes. Need formal columns (`brand`, `category`, `category_path`) to ease analytics and similarity joins.
3. **Image URLs / Feature Vectors** – Required for visual embeddings (future RT-5B). Store URLs in `props.media` and plan for vector store integration.
4. **Inventory Signals** – Stock levels or click-through on detail pages; currently absent. Optional for weighting future retrievers.
5. **Language / Locale** – For multi-language support; currently not captured.

## 3. Proposed Data Additions (Pre-RT-5B)

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| `description` | text | Shop CMS export | Use for textual embeddings; ensure sanitation. |
| `brand` | text | Normalise from tags or product feed | Replace `brand:` prefix usage. |
| `category` | text | Normalise to canonical taxonomy | Provide for analytics + retrieval. |
| `category_path` | text[] | Derived | Enables hierarchical similarity. |
| `image_url` | text | Product media service | Input for visual embedding pipeline. |
| `props.metadata_version` | text | Ingestion job | Track freshness of metadata. |

## 4. Data Pipeline Owners

- **Product ingestion**: Shop backend to emit enriched payloads; extend `/items:upsert` contract.
- **Backfill**: run one-time script (dbt or Go) to hydrate new columns from existing tags + upstream feed.
- **Validation**: add schema checks in `api/test/store` to ensure new columns populated for sample fixtures.

## 5. Next Steps

1. Update swagger/specs to include optional `brand`, `category`, `description`, `image_url` fields (pending RT-5C).
2. Coordinate with data engineering to expose product feed with the required fields.
3. Define embedding generation pipeline (tooling, batch schedule) for RT-5B.
4. Update runbooks (diversity/personalisation dashboards) to incorporate new metadata for monitoring.
