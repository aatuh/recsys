---
diataxis: explanation
tags:
  - recsys-eval
---
# Default evaluation pack (recommended)
This page explains Default evaluation pack (recommended) and how it fits into the RecSys suite.


## Who this is for

- Teams running a RecSys pilot and needing a “good enough” default metric set.
- Engineers and analysts who want to standardize ship/hold/rollback decisions across surfaces.

## What you will get

- A default metric pack you can adopt in week 1 (offline) and week 3 (online).
- A short list of guardrails that prevent “shipping on broken data”.
- Links to the canonical playbook and workflows.

## Week 1: prove the loop is real (offline + integrity)

In week 1, your goal is not to win. It’s to make measurement trustworthy.

### Must-pass integrity checks

- Schema validation passes (`recsys-eval validate`).
- Join integrity is sane (broken joins make all metrics fiction).
- Empty-recs rate is understood (and has a safe fallback UX).

Example (validate inputs):

```bash
./bin/recsys-eval validate --schema exposure.v1 --input exposures.jsonl
./bin/recsys-eval validate --schema outcome.v1 --input outcomes.jsonl
```

See:

- Decision playbook: [Decision playbook: ship / hold / rollback](decision-playbook.md)
- Suite workflow: [How-to: run evaluation and make ship decisions](../../how-to/run-eval-and-ship.md)

### Default offline metrics (regression gate)

Pick 1–2 relevance proxies and 1–2 distribution metrics:

- Relevance proxies: `hitrate@k`, `precision@k`, `ndcg@k`, `map@k`
- Distribution metrics: `coverage@k`, `novelty@k`, `diversity@k`

Start with `k=5` or `k=10` and keep it stable across runs.

Read more:

- Metrics reference: [Metrics: what we measure and why](metrics.md)
- Offline gate workflow: [Workflow: Offline gate in CI](workflows/offline-gate-in-ci.md)

## Week 3: measure impact (online experiments)

Once logging and joins are trustworthy, prefer online experiments for KPI lift.

### Default experiment metrics

- 1 primary KPI (business-owned): CTR / conversion rate / revenue per exposure (pick one)
- 2–4 guardrails (must not regress):
  - empty-recs rate
  - error rate
  - latency (p95/p99)
  - join integrity (if join-rate drops, HOLD and fix logging)

Read more:

- Online A/B workflow: [Workflow: Online A/B analysis in production](workflows/online-ab-in-production.md)
- Interpreting results: [Interpreting results: how to go from report to decision](interpreting_results.md)

## Slice keys (keep it boring)

Default slices to start with:

- `tenant_id`
- `surface`

Add one more slice only if you will act on it (device, locale, segment).

## Common pitfalls

- Shipping a “win” when join-rate is low.
- Over-slicing (finding fake wins by chance).
- Treating offline metrics as business KPIs.

## Read next

- Decision thresholds and what-to-do branches: [Decision playbook: ship / hold / rollback](decision-playbook.md)
- Suite how-to (runnable commands): [How-to: run evaluation and make ship decisions](../../how-to/run-eval-and-ship.md)
- Minimum instrumentation spec: [Minimum instrumentation spec (for credible evaluation)](../../reference/minimum-instrumentation.md)
