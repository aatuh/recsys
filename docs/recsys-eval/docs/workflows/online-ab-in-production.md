---
diataxis: explanation
tags:
  - how-to
  - evaluation
  - business
  - recsys-eval
---
# Workflow: Online A/B analysis in production
This page explains Workflow: Online A/B analysis in production and how it fits into the RecSys suite.


## Who this is for

- Product + analytics teams running experiments on key surfaces
- Engineers who need a repeatable “measure → decide → ship/rollback” workflow

## Goal

Measure impact from live traffic and decide **ship / hold / rollback** using experiment analysis.

## Prerequisites (must be true)

- You can log:
  - exposures (what was shown)
  - outcomes (what the user did)
  - assignments (experiment id + variant)
- Your join keys are stable (typically `request_id`).

Start here if anything is unclear:

- Integration logging plan: [Integration: how to produce the inputs](../integration.md)
- Data contracts: [Data contracts: what inputs look like](../data_contracts.md)

## Workflow steps

1. Pick a primary KPI and 2–4 guardrails (latency, empty-recs rate, error rate, etc.).
1. Run `recsys-eval` in `experiment` mode for a well-defined window.
1. Interpret results:
   - join-rate sanity
   - SRM (sample ratio mismatch) warnings
   - guardrails holding
1. Decide ship/hold/rollback and save the report as an audit artifact.

## Example command (experiment analysis)

```bash
recsys-eval run \
  --mode experiment \
  --dataset configs/examples/dataset.jsonl.yaml \
  --config configs/eval/experiment.default.yaml \
  --output /tmp/experiment_report.md \
  --output-format markdown
```

## Read next

- Interpretation cheat sheet: [Interpretation cheat sheet (recsys-eval)](interpretation-cheat-sheet.md)
- Interpreting results (deep dive): [Interpreting results: how to go from report to decision](../interpreting_results.md)
- Metrics: [Metrics: what we measure and why](../metrics.md)
- Troubleshooting (joins, SRM, anomalies): [Troubleshooting: symptom -> cause -> fix](../troubleshooting.md)
