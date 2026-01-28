# Recsys Shop — Epics & Tickets

## Main Description
A comprehensive, execution-ready work plan to integrate the demo “Online Shop” with Recsys as a production‑grade ecommerce recommender. It covers data modeling, catalog/user sync, event instrumentation, recommendation surfaces, experimentation, ranking/bandits, embeddings, admin/ops, security, performance, documentation and testing. Each ticket is a plain-English unit of work with a checkbox, ID, name, and outsider-friendly description so that an LLM code assistant (e.g., Cursor) can implement it without prior project context.

> Conventions: Ticket IDs are `RS-###`. Epics are `EPIC-##`. Check a box when complete.

---

## EPIC-01 — Data Modeling & Contracts
- [x] **RS-001 — Finalize Item JSON contract**  
  Define the product payload (required: `item_id`; recommended: `available`, `price`, `tags` for `brand:*` and `category:*`, and `props.*`). Document examples and edge cases (bundles, hidden, discontinued).
- [x] **RS-002 — Finalize User JSON contract**  
  Define the user payload (`user_id`, non‑PII `traits` such as locale, device, loyalty tier). Specify allowed types and validation rules.
- [x] **RS-003 — Finalize Event dictionary**  
  Lock in event types (`view=0`, `click=1`, `add=2`, `purchase=3`, `custom=4`), value semantics, required `meta` keys, and timestamp policy (UTC ISO‑8601).
- [x] **RS-004 — Recommendation-context metadata spec**  
  Standardize `meta.surface`, `widget`, `recommended`, `request_id`, `rank`, `session_id`, `referrer` to audit recs and enable A/B analysis.
- [x] **RS-005 — Field mapping DB→Recsys**  
  Map shop DB columns to Recsys fields (products → items; users → users; app events → recsys events). Include transformation notes.
- [x] **RS-006 — Validation schemas**  
  Author JSON Schemas (or Zod) for items, users, events; add CI checks and example fixtures.

---

## EPIC-02 — Catalog & Inventory Sync
- [x] **RS-007 — Item upsert pipeline**  
  Build a job/endpoint to upsert items in batches with retries and idempotency using product CRUD events.
- [x] **RS-008 — Availability synchronization**  
  Toggle `available` on stock changes and post‑checkout; ensure OOS items aren't recommended.
- [x] **RS-009 — Price update propagation**  
  Push price changes promptly; include currency and rounding rules.
- [x] **RS-010 — Tag normalization**  
  Generate consistent `brand:*`, `category:*`, optional `cat:*`, and facet tags (e.g., `color:black`).
- [x] **RS-011 — Media & URL hygiene**  
  Validate `props.url` and `props.image_url`; fallback images and canonical URLs.
- [ ] **RS-012 — Soft delete & archival**  
  Mark discontinued items and ensure recommendation exclusion without breaking history.

---

## EPIC-03 — User Sync & Identity
- [x] **RS-013 — User upsert on auth events**  
  Upsert users at signup/login with initial `traits` (locale, device, country) and last_seen timestamp.
- [ ] **RS-014 — Anonymous→Logged‑in merge**  
  Merge session history into user on login while preserving idempotency.
- [x] **RS-015 — Trait enrichment**  
  Populate loyalty tier, newsletter flag, preferred categories; define defaults.
- [x] **RS-016 — PII scrubbing**  
  Ensure no emails/phone numbers leak into `traits` or `meta`; document redaction.
- [ ] **RS-017 — Consent-aware tracking**  
  Respect tracking consent; disable or downscope metadata when not granted.
- [ ] **RS-018 — User deletion handler**  
  Implement GDPR-style delete/forget across caches and logs.

---

## EPIC-04 — Event Instrumentation (Web & API)
- [x] **RS-019 — PDP view event**  
  Emit `view` on product detail open with `meta.surface="pdp"`.
- [x] **RS-020 — Listing impression events**  
  Emit `view` per visible item in carousels/lists with `widget` and rank.
- [x] **RS-021 — Click events with context**  
  Emit `click` with `href`, `text`, `recommended`, `request_id`, `rank`.
- [x] **RS-022 — Add‑to‑cart events**  
  Emit `add` with `value=qty`, `meta.cart_id`, `meta.unit_price`, `currency`.
- [x] **RS-023 — Purchase (line‑item) events**  
  Emit one `purchase` per line item (`value=qty`, `order_id`, `unit_price`, `currency`).
- [x] **RS-024 — Order summary event (custom)**  
  Emit `custom` order‑level event **without** `item_id` with `meta.kind="order"`, `total`, `items[]`.
- [x] **RS-025 — Session & request propagation**  
  Generate `session_id` and pass through all events; capture `request_id` from recs API.
- [x] **RS-026 — Idempotency & retries**  
  Use `source_event_id`; implement retry with backoff and dedupe on server.
- [x] **RS-027 — Clock/UTC discipline**  
  Normalize timestamps to UTC with server‑side authoritative time.
- [x] **RS-028 — Batch flush service**  
  Buffer and POST `/v1/events:batch` with size/time thresholds; handle partial failures.

---

## EPIC-05 — Recsys API Client & Plumbing
- [x] **RS-029 — API client wrapper**  
  Provide typed helpers for items/users/events upserts and recommendations/similar endpoints.
- [x] **RS-030 — Namespace & env config**  
  Configure namespace, base URL, and keys via environment settings.
- [x] **RS-031 — Error taxonomy & handling**  
  Differentiate 4xx vs 5xx; implement retries/alerts only for transient errors.
- [ ] **RS-032 — Health/smoke checks**  
  Add `/readyz` probes that call a lightweight endpoint.
- [x] **RS-033 — Similar items integration**  
  Implement `/items/{id}/similar` client with fallbacks when insufficient data.

---

## EPIC-06 — Recommendation Surfaces (UI)
- [x] **RS-034 — Home: Top Picks widget**  
  Render personalized list with request/response logging and click tracking.
- [x] **RS-035 — PDP: Similar Items widget**  
  Show visually similar/complementary items; pass `include_tags_any` filters if needed.
- [ ] **RS-036 — Cart: You May Also Like**  
  Use basket context to fetch complementary items; avoid items already in cart.
- [ ] **RS-037 — Orders: Buy Again**  
  Recommend repurchase/cross‑sell items post‑checkout.
- [x] **RS-038 — Empty state fallbacks**  
  Show popular or category best‑sellers when personalization cold.
- [x] **RS-039 — Loading/error UI**  
  Graceful skeletons and retry controls; no layout shift.
- [x] **RS-040 — Accessibility & tracking**  
  Ensure keyboard/ARIA compatibility; tracking does not break a11y.

---

## EPIC-07 — Constraints, Caps & Personalization Controls
- [x] **RS-041 — Price range constraint**  
  Support `constraints.price_between` in recs requests from UI controls.
- [x] **RS-042 — Tag include/exclude**  
  Support `include_tags_any` and optional `exclude_tags_any` to steer results.
- [x] **RS-043 — Brand/category caps**  
  Implement runtime caps using consistent tagging.
- [x] **RS-044 — OOS & availability filter**  
  Exclude `available=false` items at request time.
- [x] **RS-045 — Diversity & de‑dup**  
  Avoid near‑duplicates by SKU/parent; enforce minimal brand diversity when needed.

---

## EPIC-08 — Analytics, Experimentation & Evaluation
- [x] **RS-046 — Event volume dashboard**  
  Track counts by type over time; alert on anomalies.
- [x] **RS-047 — Funnel metrics**  
  Compute CTR, ATC rate, conversion rate, revenue/user per surface.
- [x] **RS-048 — Recommendation effectiveness**  
  Attribute clicks/purchases to `recommended=true` with `request_id` and `rank`.
- [ ] **RS-049 — Offline evaluation set**  
  Build train/test splits; compute Recall@K, NDCG@K baselines.
- [ ] **RS-050 — A/B experiment framework**  
  Bucket users, assign variants, log `experiment` and `ab_bucket` in `meta`.
- [ ] **RS-051 — Experiment results reporting**  
  Sequential tests with guardrails; significance and power checks.
- [ ] **RS-052 — Data quality checks**  
  Validate required fields, schema drift, timestamp gaps; add CI job.
- [ ] **RS-053 — Reconciliation job**  
  Cross‑check purchase events vs. order DB; alert mismatches.

---

## EPIC-09 — Bandits & Ranking Tuning
- [ ] **RS-054 — Event-type weights upsert**  
  Set default weights (view 0.05, click 0.20, add 0.70, purchase 1.00, custom 0.10).
- [ ] **RS-055 — Half‑life tuning**  
  Tune recency decay by category velocity (e.g., electronics vs. books).
- [ ] **RS-056 — Explore/exploit configuration**  
  Enable/adjust bandit exploration rate per surface.
- [ ] **RS-057 — Ranking feature audit**  
  Verify price/availability/tag features pass through to ranker.
- [ ] **RS-058 — Guardrail metrics**  
  Monitor bad outcomes (OOS clicks, high returns) and add penalties if supported.

---

## EPIC-10 — Embeddings & Similarity
- [ ] **RS-059 — Text embedding pipeline**  
  Generate 384‑d embeddings from title+description+brand+category.
- [ ] **RS-060 — Embedding backfill**  
  Compute and upsert embeddings for historical catalog.
- [ ] **RS-061 — Incremental updates**  
  Recompute embeddings on product updates; schedule batch job.
- [ ] **RS-062 — Similarity validation**  
  Spot‑check nearest neighbors; blacklist false positives.
- [ ] **RS-063 — Index health checks**  
  Monitor missing/zero vectors and rebuild thresholds.

---

## EPIC-11 — Admin Tools & Backfills
- [x] **RS-064 — Event-type config admin**  
  Admin UI to view/update weights and half‑lives per namespace.
- [x] **RS-065 — Recsys re‑sync tool**  
  Admin action to resend all items/users in batches.
- [ ] **RS-066 — Historical events backfill**  
  Script to replay past events with original timestamps.
- [ ] **RS-067 — Catalog diff viewer**  
  Compare local catalog vs. Recsys snapshot; show mismatches.
- [ ] **RS-068 — Feature flags panel**  
  Toggle surfaces, caps, experiments at runtime.
- [ ] **RS-069 — Support runbook**  
  On‑call procedures for failures, rollback, and incident comms.

---

## EPIC-12 — Security, Privacy & Compliance (not needed)
- [ ] **RS-070 — Secret management**  
  Store API keys securely; rotate and scope environments.
- [ ] **RS-071 — Rate limiting & abuse controls**  
  Client/server‑side rate limits and input validation.
- [ ] **RS-072 — Privacy review**  
  Audit that no PII enters `traits`/`meta`; document DSR flows.
- [ ] **RS-073 — Data retention policy**  
  Define retention windows and deletion processes.
- [ ] **RS-074 — Access controls**  
  Restrict admin tools and analytics to authorized roles.

---

## EPIC-13 — Performance, Reliability & Cost (not needed)
- [ ] **RS-075 — Batch/concurrency tuning**  
  Tune event batch sizes and parallelism for throughput and cost.
- [ ] **RS-076 — Response caching**  
  Cache recommendation responses with short TTL; add cache keys.
- [ ] **RS-077 — SLOs & alerts**  
  Define latency/error SLOs for recs endpoint; alert on breaches.
- [ ] **RS-078 — Degradation strategies**  
  Fallback to popular items and cached results on failure.
- [ ] **RS-079 — Cost dashboard**  
  Track API call volume, compute spend proxies, and cache hit rate.
- [ ] **RS-080 — Load & chaos tests**  
  Simulate spikes and component failures; verify resilience.

---

## EPIC-14 — Documentation & Testing (not needed)
- [ ] **RS-081 — Data contract doc**  
  Publish the finalized items/users/events contract and examples.
- [ ] **RS-082 — API usage examples**  
  Provide code snippets for each endpoint; include error handling.
- [ ] **RS-083 — Onboarding guide**  
  Step‑by‑step setup for new engineers; local/test env instructions.
- [ ] **RS-084 — Unit tests for payload builders**  
  Validate item/user/event builders against schemas.
- [ ] **RS-085 — Integration tests with mock server**  
  Verify calls to Recsys endpoints and retry logic.
- [ ] **RS-086 — E2E shopping flow test**  
  Browser test that generates views/clicks/add/purchase and verifies events.
- [ ] **RS-087 — Release checklist**  
  Pre‑launch gates (monitoring, flags, rollbacks, backfills) and owner sign‑offs.
