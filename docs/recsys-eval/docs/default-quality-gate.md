---
tags:
  - evaluation
  - ops
  - recsys-eval
---

# Default quality gate contract (recommended)

## Who this is for

- Recommendation engineers standardizing “what good looks like” across surfaces.
- Engineers wiring `recsys-eval` into CI/CD as a regression gate.
- Teams that want a default policy before they have product-specific thresholds.

## What you will get

- A default gate contract (data quality + offline regression + online guardrails).
- A thresholds template you can copy and adapt.
- A CI usage example and an explicit override policy.

## The contract (what must be true)

This contract is intentionally boring. The goal is to avoid shipping on broken data or “wins” that break guardrails.

1. **Data quality must pass** (always)

- Inputs validate (`recsys-eval validate` passes for every input file you use).
- Join integrity is sane for the slices you care about (if join-rate drops, hold and fix logging).

1. **Offline regression gate must pass** (recommended in CI for every change)

- Run `offline` mode against a pinned baseline report.
- Gate on a small number of stable, interpretable metrics.

Default starting point:

- Primary: `ndcg@10` must not drop more than **0.01 absolute** versus baseline.
- Optional: add one distribution metric gate (`coverage@10`, `novelty@10`, or `diversity@10`) to catch “popularity
  collapse”.

1. **Online experiments must pass** (when you can run them)

- Pick 1 primary KPI (business-owned).
- Add 2–4 guardrails (must not regress): empty-recs rate, error rate, latency p95, and join integrity.

See the canonical decision policy (ship/hold/rollback): [`decision-playbook.md`](decision-playbook.md).

## Threshold templates (copy/paste)

### Offline gate (CI) template

```yaml
mode: offline
offline:
  metrics:
    - name: ndcg
      k: 10
    - name: coverage
      k: 10
  slice_keys:
    - tenant_id
    - surface
  gates:
    # Gate metrics use the "<name>@<k>" convention.
    - metric: ndcg@10
      max_drop: 0.01
    # Optional: prevent distribution collapse.
    - metric: coverage@10
      max_drop: 0.02
```

Notes:

- `slice_keys` refer to keys in `exposure.context` (a string map).
- Use a stricter `max_drop` (for example `0.001`) for tiny golden datasets, and a looser one for real logs.

### Experiment template (guardrails + KPI)

```yaml
mode: experiment
experiment:
  experiment_id: "exp_replace_me"
  control_variant: "control"
  primary_metrics:
    - ctr
  slice_keys:
    - tenant_id
    - surface
  guardrails:
    max_latency_p95_ms: 300
    max_error_rate: 0.01
    max_empty_rate: 0.02
```

## CI usage example (offline gate)

Recommended pattern:

1. Validate inputs (schemas).
2. Run offline evaluation with a baseline report.
3. Fail CI if gates fail (deterministically).

```bash
recsys-eval validate --schema exposure.v1 --input exposures.jsonl
recsys-eval validate --schema outcome.v1 --input outcomes.jsonl

recsys-eval run \
  --mode offline \
  --dataset configs/examples/dataset.jsonl.yaml \
  --config configs/eval/offline.ci.yaml \
  --output /tmp/offline_report.json \
  --baseline testdata/golden/offline.json
```

## Override policy (when changing thresholds is allowed)

Overrides are allowed, but they must be explicit and auditable.

Allowed reasons:

- You changed the objective intentionally (new KPI, new catalog definition, new slice key).
- You changed evaluation wiring (schema version, join logic, dataset window).
- You proved the default threshold is too strict/loose using a real report and can justify the new value.

Not allowed:

- Lowering thresholds to “make CI green” without a written reason and a linked report.
- Overriding data quality gates (broken validation or broken joins).

Recommended practice:

- Treat gate thresholds like API contracts: reviewed changes only.
- Store the effective config and the baseline report alongside the build artifact.

## Common pitfalls

- Gating on too many slices/metrics (you will find fake regressions).
- Using offline metrics as business KPI proxies without an online validation step.
- Ignoring join quality (all metrics become fiction).

## Read next

- Suite how-to (runnable commands): [`how-to/run-eval-and-ship.md`](../../how-to/run-eval-and-ship.md)
- Decision policy: [`decision-playbook.md`](decision-playbook.md)
- Offline CI workflow: [`workflows/offline-gate-in-ci.md`](workflows/offline-gate-in-ci.md)
- Default metric pack: [`default-evaluation-pack.md`](default-evaluation-pack.md)
- Metrics reference: [`metrics.md`](metrics.md)
