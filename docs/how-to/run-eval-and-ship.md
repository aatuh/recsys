---
tags:
  - how-to
  - evaluation
  - ml
  - business
  - recsys-eval
---

# How-to: run evaluation and make ship decisions

## Who this is for

- Engineers shipping recommender changes and needing a quality gate
- Analysts validating impact from logs
- Operators who need an auditable “ship / hold / rollback” decision trail

## What you will get

- A runnable baseline workflow for validating logs and producing reports
- A clear recommendation for when to use offline vs experiment analysis
- Links to the deeper `recsys-eval` docs for interpretation and scaling

## Goal

Turn exposure/outcome logs into a report you can use to decide **ship / hold / rollback**.

## Prereqs

- `recsys-eval` built (from this repo):

  ```bash
  cd recsys-eval
  make build
  ```

- Logs in the v1 schemas:
  - exposures: `exposure.v1`
  - outcomes: `outcome.v1`
  - assignments: `assignment.v1` (required for experiment mode)

## 0) Validate inputs (always)

Validation is strict (extra fields can fail). Run this before trusting any metric:

```bash
./bin/recsys-eval validate --schema exposure.v1 --input exposures.jsonl
./bin/recsys-eval validate --schema outcome.v1 --input outcomes.jsonl
./bin/recsys-eval validate --schema assignment.v1 --input assignments.jsonl
```

Tip: if you want `recsys-service` to emit `exposure.v1` directly, set:

- `EXPOSURE_LOG_ENABLED=true`
- `EXPOSURE_LOG_FORMAT=eval_v1`

See: [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)

## 1) Run an offline regression gate (recommended baseline)

Always run an offline regression gate in CI:

- compare baseline vs candidate versions
- fail if a primary metric regresses beyond a threshold

Example:

```bash
./bin/recsys-eval run \
  --mode offline \
  --dataset configs/examples/dataset.jsonl.yaml \
  --config configs/eval/offline.ci.yaml \
  --output /tmp/offline_report.md \
  --output-format markdown
```

If your exposure logs come from `recsys-service` in `eval_v1` format, the exposure `context` keys are named like
`tenant_id`, `surface`, and `segment`. Ensure your `slice_keys` use the same names.

## 2) Prefer online experiments when possible

Online A/B tests are the best way to measure real impact:

- log exposures with experiment id/variant
- log outcomes tied to the same `request_id`
- check KPI + guardrails

Example:

```bash
./bin/recsys-eval run \
  --mode experiment \
  --dataset configs/examples/dataset.jsonl.yaml \
  --config configs/eval/experiment.default.yaml \
  --output /tmp/experiment_report.md \
  --output-format markdown
```

## 3) Ship / rollback mechanics

Ship if KPI improves and guardrails hold. Hold if results are inconclusive. Roll back if primary or guardrails regress.

Rollback levers:

- Artifacts/manifest: swap the manifest pointer (pipelines) and invalidate service caches
- Config/rules: roll back config/rules versions and invalidate service caches

See:

- Pipelines rollback: [`recsys-pipelines/docs/how-to/rollback-manifest.md`](../recsys-pipelines/docs/how-to/rollback-manifest.md)
- Admin cache invalidation: [`reference/api/admin.md`](../reference/api/admin.md)
- Interpreting results: [`recsys-eval/docs/interpreting_results.md`](../recsys-eval/docs/interpreting_results.md)

## Verify

- The report file exists and includes a summary table for your chosen mode.
- Join integrity is sane (if join rate is low, fix logging before trusting metrics).

## Read next

- Decision playbook (ship / hold / rollback): [`recsys-eval/docs/decision-playbook.md`](../recsys-eval/docs/decision-playbook.md)
- Experimentation model: [`explanation/experimentation-model.md`](../explanation/experimentation-model.md)
- Exposure logging and attribution: [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)
