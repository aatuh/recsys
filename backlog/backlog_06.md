# Audit Trail of Decisions — README

## TL;DR

Add an append-only, structured **DecisionTrace** for each (or sampled)
recommendation response. Capture inputs → effective config → ranking steps →
final output and reasons. Persist asynchronously (to Postgres or a stream).
Expose read APIs and an admin UI. This yields compliance, debuggability, and
clear KPI attribution with minimal runtime cost.

---

## Goals

* **Compliance & trust**: Show exactly why items were recommended.
* **Debuggability**: Reproduce decisions and investigate discrepancies.
* **Attribution**: Tie outcomes (clicks, plays, revenue) to recommendations.
* **Low overhead**: Asynchronous writes; configurable sampling; bounded
  storage.

---

## What to capture per decision

Store a single JSON object, `DecisionTrace`, with these sections:

1. **Request meta**

   * `decision_id` (UUID), `ts` (UTC)
   * `namespace`, `surface` (optional), `request_id`
   * `user_hash` (salted hash of user\_id; no raw PII)
   * `k` (result size), constraints (e.g., `exclude_item_ids`)

2. **Effective config** (the knobs used for this request)

   * Blend weights: `alpha` (pop), `beta` (co-vis), `gamma` (embed)
   * Personalization: `profile_boost`, `profile_window_days`, `profile_top_n`
   * Diversity & caps: `mmr_lambda`, `brand_cap`, `category_cap`
   * Windows & rules: `half_life_days`, `co_vis_window_days`,
     `purchased_window_days`, `rule_exclude_purchased`
   * Segment/Profile IDs if segmenting is enabled

3. **Bandit context** (if used)

   * `chosen_policy_id`, `algorithm`, `bucket_key`, `explore` (bool)
   * Optional `bandit_explain`

4. **Ranking stages**

   * `candidates_pre`: array of `{item_id, score}` before MMR/caps
   * `mmr_info`: optional details per pick (e.g., `max_sim`, cap hits)
   * `final_items`: array of `{item_id, score, reasons[]}` after MMR/caps

5. **Per-item reasons**

   * Deterministic tags such as `recent_popularity`, `co_visitation`,
     `embedding_similarity`, `personalization`, `diversity`, `cap_applied`.

### Example `DecisionTrace`

```json
{
  "decision_id": "0c5f1a95-4b8d-4f63-97b4-3f8b43b1f3c0",
  "ts": "2025-09-18T06:05:12Z",
  "namespace": "casino-fi",
  "surface": "home",
  "request_id": "req_01HXR...",
  "user_hash": "b4e2c5...",
  "k": 10,
  "constraints": {"exclude_item_ids": ["slot_999"]},
  "effective_config": {
    "alpha": 0.7, "beta": 0.2, "gamma": 0.1,
    "profile_boost": 0.2, "profile_window_days": 30, "profile_top_n": 12,
    "mmr_lambda": 0.6, "brand_cap": 2, "category_cap": 3,
    "half_life_days": 3, "co_vis_window_days": 14,
    "purchased_window_days": 7, "rule_exclude_purchased": true,
    "segment_id": "vip", "profile_id": "vip-high-novelty"
  },
  "bandit": {
    "chosen_policy_id": "blend-0.7-0.2-0.1",
    "algorithm": "thompson_beta",
    "bucket_key": "home",
    "explore": false
  },
  "candidates_pre": [
    {"item_id": "slot_42", "score": 0.8129},
    {"item_id": "slot_17", "score": 0.7811}
  ],
  "mmr_info": [
    {"pick_index": 0, "item_id": "slot_42", "max_sim": 0.33,
     "brand_cap_hit": false, "category_cap_hit": false}
  ],
  "final_items": [
    {"item_id": "slot_42", "score": 0.813,
     "reasons": ["recent_popularity", "co_visitation", "personalization",
                  "diversity"]},
    {"item_id": "slot_08", "score": 0.699,
     "reasons": ["embedding_similarity", "diversity"]}
  ],
  "extras": {
    "anchors": ["slot_17", "slot_08"],
    "explain_level": "numeric"
  }
}
```

---

## Storage (Postgres, JSONB)

Two tables (partitioned by day if volume is high):

```sql
CREATE TABLE rec_decisions (
  decision_id    UUID PRIMARY KEY,
  ts             TIMESTAMPTZ NOT NULL,
  namespace      TEXT NOT NULL,
  surface        TEXT,
  request_id     TEXT,
  user_hash      TEXT,
  effective_config JSONB NOT NULL,
  bandit         JSONB,
  candidates_pre JSONB NOT NULL,
  final_items    JSONB NOT NULL,
  extras         JSONB,
  -- indexes
  CONSTRAINT rec_decisions_ts CHECK (ts IS NOT NULL)
);
CREATE INDEX idx_recdec_ns_ts ON rec_decisions (namespace, ts);
CREATE INDEX idx_recdec_req ON rec_decisions (request_id);
CREATE INDEX idx_recdec_user_ts ON rec_decisions (user_hash, ts);

-- Optional normalized table for per-pick mmr/caps insights
CREATE TABLE rec_decisions_mmr (
  decision_id    UUID NOT NULL,
  pick_index     INT NOT NULL,
  item_id        TEXT NOT NULL,
  max_sim        DOUBLE PRECISION,
  brand_cap_hit  BOOLEAN,
  category_cap_hit BOOLEAN,
  PRIMARY KEY (decision_id, pick_index)
);
```

JSONB allows easy evolution; only index what you query frequently.

---

## Write path (server)

1. After ranking and before responding, assemble `DecisionTrace` from locals.
2. Push it onto a bounded channel/queue.
3. An async worker batches inserts every N ms or when batch size is reached.
4. On DB error: retry with backoff; optionally spill to disk (rotating files)
   and drop oldest beyond a size budget.

### Sampling & retention

* Sampling per namespace (e.g., 10–100%). Log all for critical surfaces.
* Retention policy: 14–30 days hot in Postgres; optional archive to object
  storage as NDJSON. TTL per namespace.

---

## Read APIs (admin)

* `GET /v1/audit/decisions?namespace=...&from=...&to=...&user_hash=...&request_id=...`

  * Paginated list (metadata + final\_items top-K).
* `GET /v1/audit/decisions/{decision_id}`

  * Full DecisionTrace.
* `POST /v1/audit/search`

  * JSON body for advanced filters (reason contains, policy, surface, etc.).

---

## Admin UI

* **Audit list**: time, namespace, surface, policy, top item previews.
* **Filters**: time range, reason contains, policy id, user hash, request id.
* **Detail modal**: request meta, effective config, candidates\_pre vs final,
  reasons chips per item, mmr/caps/bandit sections.
* **User journey (optional)**: stitch events + rec snapshots for exploration.

---

## Privacy & compliance

* Store only a salted hash of user\_id:

```text
user_hash = HEX(SHA256(namespace || ":" || user_id || ":" || secret_salt))
```

* Never store IPs, user agents, or free-text PII.
* If needed, mask geo to region codes; store item IDs (resolve display names
  in UI via catalog lookups).

---

## Performance & reliability

* Async writer keeps p95 latency flat; make sampling configurable.
* Batch inserts (e.g., 100–500 records) for throughput.
* Back-pressure: if the queue is full, drop to sampling floor and log.
* Health metrics: queue depth, insert rate, failures, spill-to-disk bytes.

---

## Testing

* **Golden test**: call recommendations with a deterministic fixture; fetch the
  corresponding audit record; assert ids, scores, and reasons match the HTTP
  response exactly.
* **MMR/caps assertions**: if MMR is enabled, assert expected differences
  between `candidates_pre` and `final_items`.
* **Load test**: verify async writer maintains latency budgets (e.g., p95 ≤
  150ms) under sustained RPS with sampling on.

---

## Implementation checklist

* [x] Define `DecisionTrace` struct and JSON marshaling.
* [x] Hook capture in the recommend handler (post-rank, pre-response).
* [ ] Async writer with batching, retry, spill-to-disk, and metrics.
* [x] Postgres DDL + migrations; optional partitioning.
* [x] Read endpoints (`GET`/`POST`) and auth.
* [ ] Admin UI pages (list + detail) reusing existing item/reason components.
* [ ] Configurable sampling & retention by namespace.
* [ ] Golden + load tests; dashboards for writer health.

---

## Notes

* Keep audit writes best-effort. The recommendation response must never wait on
  audit persistence.
* Store exactly what the user saw; do not recompute when reading.
* Prefer compact arrays of `{item_id, score, reasons[]}` for final items; keep
  large stage internals optional to control size.
