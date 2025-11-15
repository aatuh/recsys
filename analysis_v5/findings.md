# Main Findings – recsys Evaluation (v5)

## Context
- **Namespaces evaluated:** `retail_en_us` (commerce, en-US) and `media_fi_fi` (streaming, fi-FI), seeded via `analysis_v5/sql/seed_eval.sql`.
- **Tools used:** `analysis/scripts/run_quality_eval.py`, direct `/v1/recommendations` & `/v1/admin/*` calls, Prometheus `/metrics`, and trace evidence under `analysis/evidence`.
- **Baseline:** Popularity-only ranking.
- **Model version:** `popularity_v1` (per recommendation responses).

---

## 1. Recommendation Quality & Coverage
- **Warm users:** Offline lift vs. popularity baseline remains strong (≈+133% NDCG in retail, +341% in media; see `analysis_v5/results/retail_quality.json:1-70` and `media_quality.json:1-90`). Catalog coverage stays at 100% of the tiny seed, and long-tail share ≥0.66.
- **Limitations:** Evaluation touches only 4 users per namespace—far below the 100-user requirement—so metrics lack statistical confidence. `/v1/items/{id}/similar` returns empty arrays because embeddings are never populated and there is no ANN store (`analysis_v5/sql/seed_eval.sql:32-110`). `/rerank` is entirely missing from the API surface (`api/specs/endpoints/endpoints.go:20-42`), so A2–A3 tests in the rubric cannot run.
- **Cold start:** Zero-data users all receive the same popularity list (e.g., `analysis_v5/results/recommendation_dump.json:34-120`), indicating the starter profile needs diversity controls. Few-shot users show reasonable lift but still rely heavily on popularity anchors.

## 2. Configurability & Policy Controls
- **Rules engine:** With `RULES_ENABLE=true` (api/.env:45) the service now enforces block/boost rules. Evidence: `analysis_v5/results/rules_block_sample.json#L1` shows `rule_block_count:1`, and manual overrides inject `rule.boost:+0.08[...]` reasons (`analysis_v5/results/rules_effect_sample.json:1-60`). Cancelling overrides removes the boost with immediate effect.
- **Remaining gaps:** Recommendation constraints only honor `include_tags_any`; `price_between` and `created_after` filters are ignored (`api/internal/algorithm/engine.go:817-878`). This causes all hard-filter tests (age gating, inventory) to fail. There is still no `/v1/admin/recommendation/config` endpoint or tooling to inspect/roll back runtime knobs—operators must edit `.env` and redeploy.

## 3. Serving Performance & Reliability
- **Current data:** Only dozens of requests were issued during evaluation. `/metrics` shows `http_request_duration_seconds` buckets up to 0.05s for recommendations, but no load/chaos testing was performed. `/version` endpoint is missing, returning 400/405 errors (see `curl` output earlier), so determinism/version tracking cannot be verified.
- **Operational tooling:** No load or failover harness, no config snapshot/rollback workflow, and no bandit experiments exercised beyond env defaults. These deficiencies keep axis D (“Serving & Reliability”) in FAIL status.

## 4. Explainability & Debugging
- **Traces:** `/v1/recommendations` responses include candidate source counts, starter profile weights, and policy summaries (e.g., `analysis/evidence/recommendation_samples_after_seed.json:2-60`). This satisfies basic explainability.
- **LLM Explain:** `/v1/explain/llm` responds with canned “llm_disabled” warnings because `LLM_EXPLAIN_ENABLED=false` (api/.env:58-65). Until a real provider is wired up, the explainability rubric remains partially unmet.

## 5. Safety, Compliance, Fairness
- **Brand/category exposure:** Analysis of all user calls (`analysis_v5/results/recommendation_dump.json:1-120`) shows brand exposure ratios up to 1.8× the mean (Voltify, NordicSignal), exceeding typical parity thresholds. No remediation (MMR tuning, post-ranking diversity) is in place.
- **Policy violations:** Price-gated requests still return high-price items because `price_between` is ignored, leading to hard-filter violations. No guardrails prevent sensitive categories from leaking when constraint logic is bypassed.
- **Auditability:** Decision traces are recorded, but there is no automated monitor for constraint leaks or rule zero-effect detection beyond log lines.

## 6. Developer Experience & Documentation
- **Docs:** README and `docs/api_endpoints.md` cover existing endpoints, but rerank/similar gaps and rule-testing workflows are undocumented. No executive-summary template exists for sharing findings with stakeholders, and evaluation artifacts are scattered across directories without a single bundle.
- **DX blockers:** Lack of `/v1/admin/recommendation/config` prevents dynamic tuning; there’s no CLI to run policy tests, and the evaluation scripts still assume “default” namespace unless overridden manually.

---

## Summary Verdict
Even after re-enabling rules, the system **fails** the production rubric because:
1. Missing capabilities (similar, rerank) block fundamental tests.
2. Constraint enforcement is incomplete, so hard safety guarantees can’t be validated.
3. Serving/ops readiness is unproven (no load tests, no versioning, no rollback tooling).
4. Fairness metrics exceed acceptable thresholds with no mitigation plan.
5. Explainability/dx tooling is partial (LLM disabled, no config API).

Refer to `analysis_v5/backlog.md` for the prioritized remediation plan.
