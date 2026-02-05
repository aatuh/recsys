---
tags:
  - how-to
  - evaluation
  - ops
  - recsys-eval
---

# Workflow: Offline gate in CI

## Who this is for

- Engineers wiring `recsys-eval` into CI/CD as a quality gate
- Teams that want an auditable “ship / hold / rollback” decision trail

## Goal

Fail builds when recommendation quality regresses beyond agreed thresholds, using an offline regression report.

## The workflow (recommended baseline)

This is the simplest reliable pattern:

1. Validate inputs (schemas + joins).
1. Run `recsys-eval` in `offline` mode.
1. Attach the report to the build (artifact).
1. Fail CI when gates fail (deterministically).

## Inputs you need

- A dataset config (what files to read and how to join them)
- An evaluation config (metrics, slices, thresholds)
- A baseline report (committed “golden” or a pinned prior run)

## Example command (CI gate)

```bash
recsys-eval run \
  --mode offline \
  --dataset configs/examples/dataset.jsonl.yaml \
  --config configs/eval/offline.ci.yaml \
  --output /tmp/offline_report.json \
  --baseline testdata/golden/offline.json
```

## Practical gating guidance

- Use **tiny “golden” datasets** for behavior regression tests (fast, stable).
- Use **real logs** for scheduled production gates (high signal).
- Treat “invalid input” differently from “metric regression”:
  - fix logging before trusting metrics
  - keep gates deterministic and auditable

## Read next

- CI gates (details + exit codes): [`recsys-eval/docs/ci_gates.md`](../ci_gates.md)
- Metrics reference: [`recsys-eval/docs/metrics.md`](../metrics.md)
- Interpreting results: [`recsys-eval/docs/interpreting_results.md`](../interpreting_results.md)
- Suite workflow: [`how-to/run-eval-and-ship.md`](../../../how-to/run-eval-and-ship.md)
