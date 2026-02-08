---
diataxis: explanation
tags:
  - evaluation
  - recsys-eval
  - workflows
---
# recsys-eval docs

recsys-eval turns recommendation logs into **evaluation reports** so you can make
clear decisions: **ship / hold / rollback**.

If you only read one page first: **Concepts** → [Concepts: how to understand recsys-eval](concepts.md)

## Choose your path

### I’m new and need the big picture

- Overview: [recsys-eval](../overview.md)
- Concepts: [Concepts: how to understand recsys-eval](concepts.md)
- Workflows:
  - Offline gate in CI: [Workflow: Offline gate in CI](workflows/offline-gate-in-ci.md)
  - Online A/B in production: [Workflow: Online A/B analysis in production](workflows/online-ab-in-production.md)

### I’m integrating data/logs

- Data contracts: [Data contracts: what inputs look like](data_contracts.md)
- Integration guide: [Integration: how to produce the inputs](integration.md)

### I’m interpreting results

- Metrics reference: [Metrics: what we measure and why](metrics.md)
- Interpretation guide: [Interpreting results: how to go from report to decision](interpreting_results.md)
- Interpretation cheat sheet: [Interpretation cheat sheet (recsys-eval)](workflows/interpretation-cheat-sheet.md)

### I’m running this in CI or on-call

- CI gates: [CI gates: using recsys-eval in automation](ci_gates.md)
- Scaling: [Scaling: large datasets and performance](scaling.md)
- Runbooks: [Runbooks: operating recsys-eval](runbooks.md)
- Troubleshooting: [Troubleshooting: symptom -> cause -> fix](troubleshooting.md)

### I’m doing a deeper evaluation method

- OPE (off-policy evaluation): [Off-policy evaluation (OPE): powerful and easy to misuse](ope.md)
- Interleaving: [Interleaving: fast ranker comparison on the same traffic](interleaving.md)
- Architecture: [Architecture: how the code is organized and how to extend it](architecture.md)

## Read next

- Suite workflow: [How-to: run evaluation and make ship decisions](../../how-to/run-eval-and-ship.md)
- Evaluation modes: [Evaluation modes](../../explanation/evaluation-modes.md)
