---
diataxis: explanation
tags:
  - reference
  - evaluation
  - business
  - recsys-eval
---
# Interpretation cheat sheet (recsys-eval)
This page explains Interpretation cheat sheet (recsys-eval) and how it fits into the RecSys suite.


## Before trusting any metric

- **Validate schemas** (extra/missing fields can break joins):
  [Data contracts: what inputs look like](../data_contracts.md)
- **Check join integrity**:
  - low match rate usually means broken instrumentation, not a “bad model”
  - fix logging before debating metric moves
- **Look for SRM warnings in experiments**:
  - SRM often indicates broken bucketing or assignment logging
  - do not ship based on experiment results with SRM you can’t explain

## If the primary KPI moved

Ask “is the move real, safe, and attributable?” in this order:

1. **Real:** enough samples, stable joins, no obvious data anomalies.
1. **Safe:** guardrails hold (latency, errors, empty recs, diversity constraints).
1. **Attributable:** change is consistent across slices you care about.

## Common “this looks wrong” signals

- KPI jumps by an impossible amount (often join issues or double-counting).
- Slice results contradict global results (often logging/slicing mismatch).
- High variance and no clear direction (often not enough traffic).

## Read next

- Interpreting results: [Interpreting results: how to go from report to decision](../interpreting_results.md)
- Runbooks (common failure modes): [Runbooks: operating recsys-eval](../runbooks.md)
- Troubleshooting: [Troubleshooting: symptom -> cause -> fix](../troubleshooting.md)
