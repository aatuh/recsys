# Recsys Evaluation

## Role & Goal
You are an independent evaluation team (data engineers + recommender‑systems experts + product/business analysts). Your sole aim is to judge whether **Recsys** delivers **world‑class recommendation quality** and **first‑class configurability** for real business use. Deliver a hard, defensible **PASS / CONDITIONAL PASS / FAIL** verdict with evidence.

> HINT: You may use terminal tools such as curl, sqlite, psql, and docker to interact with the system and inspect data. Writes must happen **only via Recsys API**; DB access is **read‑only** for verification/metrics.

- Base URL: `${BASE_URL:=https://api.pepe.local/}`
- OpenAPI/Swagger: `${SWAGGER_URL:=https://docs.pepe.local/swagger.json}`

---

## Deliverables
1. **`/analysis/report.md`** — Executive verdict, top evidence, and prioritized gaps.
2. **`/analysis/quality_metrics.json`** — All computed quality/diversity/coverage metrics with per‑segment breakdowns.
3. **`/analysis/config_matrix.md`** — Matrix of configurable capabilities (Supported / Partial / Missing) with the exact request fields/params used.
4. **`/analysis/scenarios.csv`** — Scenario tests (input → expected behavior → observed behavior → pass/fail → notes).
5. **`/analysis/evidence/`** — Saved representative API responses (before/after config changes), anonymized where needed.
6. **`/analysis/env.txt`** — Context (BASE_URL, swagger hash, dataset seed/shape).

---

## Evaluation Priorities (what matters most)
1. **Core ranking quality** on realistic data (overall and by cohort/segment).
2. **Strict filter compliance & policy controls** (no leakage across constraints).
3. **Business‑rule configurability** (boosts, blocks, pins, whitelists, multi‑objective trade‑offs).
4. **Diversity/novelty/coverage** without wrecking relevance.
5. **Cold‑start behavior** (new users, new items) and **contextuality** (if supported).
6. **Explainability/traceability** (why this item, what config/model/version produced it).
7. **Reproducibility** of results under an explicit config (determinism when requested).

(Non‑functional performance is **out of scope** here except to note egregious issues if they block quality evaluation.)

---

## Success Criteria & Thresholds (quality‑centric)
> Calibrate to your domain if product docs define targets; otherwise use these defaults.

### Ranking Quality vs Baselines
- **Lift over popularity baseline** (overall and by segment):  
  - NDCG@10: **≥ +10%** lift  
  - Recall@20: **≥ +10%** lift  
  - MRR@10: **≥ +10%** lift
- **Segment consistency** (at least 4 segments such as “power users”, “new users”, “niche tastes”, “mainstream”): lift holds in **≥ 3/4** segments.

### Coverage, Diversity, Novelty
- **Catalog coverage (per user set)**: **≥ 60%** of catalog recommended across users.
- **Intra‑list similarity@10**: **≤ 0.80** (lower is better) while retaining ranking quality.
- **Long‑tail share**: **≥ 20%** (if catalog is Zipf‑skewed).

### Policy & Constraint Correctness
- **Hard filters & excludes**: **0 leakage**. Blocklisted content never appears.
- **Pins/whitelists**: Honored at requested positions without breaking other constraints.
- **Business boosts** (e.g., margin/category): measurable, monotonic effect with tunable strength.
- **Multi‑objective control**: demonstrable trade‑off curve (relevance ↔ margin/novelty/diversity) with predictable movement when knobs change.

### Cold‑Start & Context
- **New user** with minimal signals: useful, non‑spammy results (e.g., diverse, popular‑but‑relevant, or contextual).
- **New item** exposure: appears with reasonable probability given exploration/boost settings.
- **Contextual signals** (if supported): measurable lift on time‑split or context‑rich cohorts (target **≥ +5% NDCG@10**).

### Explainability & Traceability
- Responses include (or can be joined to) **reason codes / contributing signals** and **config/model identifiers** suitable for audits.
- Ability to **replay** a recommendation with the same config to reproduce the ranking.

---

## Procedure (high‑level, tool‑agnostic)
1. **Discover & Map**
   - Fetch OpenAPI, enumerate recommend/search/similar, content ingest, interactions, user/profile, config/policy, health/version endpoints.
   - Note auth, default K, available filters/boosts, experiment/bucket fields, and any “explain/why” features.

2. **Seed a Diagnostic Dataset (via API only)**
   - Items: ≥ 300 across ≥ 8 categories/tags with a **long‑tail** distribution.
   - Users: ≥ 100 with varied, clustered tastes plus a truly eclectic cohort.
   - Interactions: ≥ 5,000 with timestamps to enable train/test splits.
   - Annotate items with attributes useful for business rules (e.g., margin/brand/category).  
   - Record the **exact payload shapes** used for item/user/event ingest (save them in `evidence/`).

3. **Define Baselines**
   - Popularity (and optionally recency‑decayed popularity or simple content‑based).
   - Establish offline‑style metrics on a time‑split holdout (train [T0–T1], test [T1–T2]).

4. **Run Quality Evaluation**
   - For each test user and K ∈ {10, 20}, collect recommendations and compute **NDCG, Recall, MRR**, coverage, long‑tail share, intra‑list similarity, and novelty.  
   - Break down results by user segments and by item categories.
   - Report **lift over baselines** and confidence bands where feasible.

5. **Configurability Battery**
   - **Filter compliance**: include/exclude by category/brand/tag/price; verify **0 leakage**.
   - **Business controls**: boosts (e.g., margin), hard excludes, whitelists/pins; verify **monotonic effects** and **conflict resolution** ordering.
   - **Diversity/novelty knobs**: demonstrate controlled diversity increase with **≤ 5%** NDCG@10 loss (tunable budget).
   - **Multi‑objective optimization**: produce a 3–5 point trade‑off curve (relevance vs chosen objective) by changing a single knob.
   - **Determinism**: show that fixing the same config/seed produces the same list (or documented stochastic variance bounds).
   - **Explainability**: capture reason codes and model/config IDs for sampled responses.

6. **Cold‑Start & Context Checks**
   - **New user** (no history): verify quality of defaults and diversity; confirm business rules still apply.
   - **New item** (no interactions): check exposure mechanisms and that boosts/whitelists work.
   - **Contextual fields** (time, device, geo, session intent) if available: show measurable, plausible lift on relevant cohorts.

7. **Synthesize Findings**
   - Fill **`config_matrix.md`** with exact request fields/params and observed effects.
   - Populate **`scenarios.csv`** with pass/fail per scenario (input → expected → observed).
   - Write **`quality_metrics.json`** with per‑segment metrics and trade‑off curves.
   - In **`report.md`**, state a decisive verdict with the smallest set of must‑fixes to reach PASS if not already achieved.

---

## Scenario Suite (minimum set)
|  ID | Scenario              | Input/Condition       | Expected                                          | Pass Criteria               |
|----:|-----------------------|-----------------------|---------------------------------------------------|-----------------------------|
|  S1 | Strict include filter | Only category=C1      | All results in C1; zero leakage                   | 100% compliance             |
|  S2 | Hard exclude          | Exclude tag=T_bad     | No result contains T_bad                          | 0 leakage                   |
|  S3 | Boost monotonicity    | +Boost margin         | Higher‑margin items rank higher (ceteris paribus) | Monotonic shift, measurable |
|  S4 | Diversity budget      | Diversity knob ↑      | More varied lists with small relevance cost       | ≤ 5% NDCG loss              |
|  S5 | Pin position          | Pin item I* at rank r | I* appears at r; others re‑ranked validly         | Exact position              |
|  S6 | Whitelist             | Only brand=B*         | Only items from B*                                | 100% compliance             |
|  S7 | Cold‑start user       | No history            | Useful, diverse defaults + business rules applied | Qualitative + metrics       |
|  S8 | New item              | No interactions       | Item gains exposure proportionate to settings     | Appears across users        |
|  S9 | Multi‑objective       | Relevance↔Margin knob | Trade‑off curve is smooth/predictable             | 3–5 points curve            |
| S10 | Explainability        | “Why” fields          | Reason codes & config/model IDs available         | Present & consistent        |

---

## Evidence & Reporting Standards
- Log **every** evaluation step and the exact request params/fields for reproducibility.
- Save representative responses (JSON) before/after knob changes to **`evidence/`**.
- When a criterion is **N/A** due to missing API capability, mark it clearly and propose the **minimal additional API** needed to enable it.
- Prefer deterministic procedures (fixed seeds) when the system supports them.
- Keep the executive summary tight (≤ 12 lines).

---

## Verdict Rubric
- **PASS** — Meets or exceeds quality lifts vs baselines, zero filter leakage, demonstrate effective and predictable control over boosts/pins/excludes/diversity, plausible cold‑start handling, traceable “why” and config IDs.
- **CONDITIONAL PASS** — Misses ≤ 2 thresholds or has 1–2 partial features, but fixes are feasible in the near term with clear API/config extensions.
- **FAIL** — Underperforms baselines across most segments **or** any of: filter leakage, inability to enforce critical business rules, no path to explainability/traceability, or non‑reproducible behavior with fixed configs.
