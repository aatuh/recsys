# Executive Summary – `<customer_or_run>`

> Replace the placeholders below after each evaluation run. Keep the entire summary to one page (≈12 lines of prose + tables) so stakeholders can digest it quickly.

- **Date:** `<YYYY-MM-DD>`
- **Org / Namespace:** `<org_id>` / `<namespace>`
- **Run / Bundle:** `analysis/reports/<customer>/<timestamp>/`
- **Scenario & Quality Status:** `<PASS/FAIL>` (reference `analysis/quality_metrics.json` & `analysis/evidence/scenario_summary.json`)

## Snapshot

| Metric                              | Value / Delta | Source                                   |
|-------------------------------------|---------------|------------------------------------------|
| Overall NDCG@10 / MRR@10 lift       | `<+xx% / +yy%>` | `analysis/quality_metrics.json`           |
| Coverage / Long-tail share          | `<0.xx / 0.yy>` | `analysis/quality_metrics.json`           |
| Cold-start S7 avg MRR / categories  | `<0.xx / n>`  | `analysis/evidence/scenario_summary.json` |
| Exposure max / mean ratio (by ns)   | `<1.xx>`      | `analysis/results/exposure_dashboard.json` |
| Determinism & Load status           | `<PASS/FAIL>` | `analysis/results/determinism_ci.json`, `analysis/results/load_test_summary.json` |

## 3 Strengths

1. `<Strength #1 — impact, metric, evidence path>`
2. `<Strength #2>`
3. `<Strength #3>`

## 3 Blockers

1. `<Blocker #1 — risk, affected surface, guardrail breached>`
2. `<Blocker #2>`
3. `<Blocker #3>`

## 3 Fast Wins

1. `<Fast win #1 — owner + ETA>`
2. `<Fast win #2>`
3. `<Fast win #3>`

### Next Actions & Owners

| Item                                 | Owner | Due date | Evidence to update |
|--------------------------------------|-------|----------|--------------------|
| `<e.g., Tune trend_seekers blend>`   | `<@>` | `<date>` | `analysis/results/...` |
| `<Add rerank scenario coverage>`     | `<@>` | `<date>` | `analysis/reports/...` |

> Attach this page to your evaluation bundle (e.g., `analysis/reports/<customer>/<timestamp>/executive_summary.md`) and link it in `analysis/findings.md` for long-term tracking.
