---
diataxis: explanation
tags:
  - explanation
  - evaluation
  - ml
  - business
  - developer
---
# Evaluation reasoning and pitfalls

This page explains how to interpret RecSys evaluation results without fooling yourself.

## Who this is for

- Recommendation engineers and analysts
- Leads approving ship/hold/rollback decisions
- Stakeholders who want to understand *why* a metric changed

## What you will get

- A checklist for sanity-checking offline reports
- The most common pitfalls that make metrics lie
- How to connect evaluation output to an operational decision

## Mental model

Evaluation answers a narrow question:

> "Given the data we logged, do we have evidence that this change helps the user and the business, without breaking safety/guardrail constraints?"

Your confidence depends on:

- data quality (joins, missingness, bias)
- methodological fit (offline vs online vs OPE)
- statistical robustness (variance, multiple testing)
- operational realism (does production match the evaluation assumptions?)

## The minimum sanity checks

Before you trust any number, confirm these:

1. **Join rate is stable**
   - Exposures join to outcomes by `request_id` at a stable rate over time.
2. **Slice coverage is stable**
   - `tenant_id`, `surface`, and your main segmentation keys exist for the same share of events.
3. **Traffic mix is comparable**
   - user cohorts and surfaces are not silently changing between runs.
4. **Guardrails are present**
   - at least one guardrail metric is reported (e.g., latency, churn proxy, diversity constraint).

See also: [Event join logic (exposures ↔ outcomes ↔ assignments)](../reference/data-contracts/join-logic.md)

## Common pitfalls

### 1) Clicks without exposures

If you log outcomes but not exposures, you cannot attribute.

**Symptom:** metrics swing wildly or look implausibly good.

**Fix:** implement exposure logging first.

- Concept: [Exposure logging and attribution](exposure-logging-and-attribution.md)
- Schemas: [Exposure, outcome, and assignment schemas](../reference/data-contracts/exposure-outcome-assignment.md)

### 2) Simpson’s paradox (your aggregate lies)

Overall lift can hide regressions in key segments.

**Fix:** require a small slice set (surface, tenant, segment) in every report and treat big slices as first-class.

### 3) Leakage and look-ahead

Offline evaluation can accidentally use information that was not available at serving time.

**Fix:** enforce time windows and event ordering in pipelines; document and test it.

- Pipelines: [Windows and backfills](../recsys-pipelines/docs/explanation/windows-and-backfills.md)

### 4) Non-stationarity

User intent, inventory, and seasonality shift.

**Fix:** keep rolling baselines; treat changes as local decisions; prefer online tests for high-impact changes.

### 5) Metric gaming

A change can improve a proxy metric while harming the actual user outcome.

**Fix:** pair every primary KPI with at least one guardrail. Prefer metrics that measure user value directly.

## Choosing evaluation mode

- **Offline regression**: fastest and cheapest; good for "did we break anything" gates.
- **Online A/B**: highest credibility for product outcomes; requires experiment controls.
- **OPE / counterfactual**: useful when A/B is expensive; sensitive to modeling assumptions.
- **Interleaving**: efficient comparisons for ranking changes; still needs careful logging.

See: [Evaluation modes](evaluation-modes.md)

## Turning a report into a decision

A good decision record answers:

- what changed
- what improved and what worsened (including slices)
- whether guardrails stayed within bounds
- ship/hold/rollback decision and why

Start here:

- [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)
- [Decision playbook (ship/hold/rollback)](../recsys-eval/docs/decision-playbook.md)

## Read next

- Evaluation modes: [Evaluation modes](evaluation-modes.md)
- How-to: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)
- Data join rules: [Event join logic (exposures ↔ outcomes ↔ assignments)](../reference/data-contracts/join-logic.md)
