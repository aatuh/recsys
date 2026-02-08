---
diataxis: explanation
tags:
  - persona
  - ml
  - recsys
  - ranking
  - evaluation
---
# Recommendation systems expert

Use this page to quickly judge the ranking/evaluation model, determinism, and the suite’s change-validation workflow.

## What you try to do in the first 10 minutes

- Understand the candidate → ranking → response flow and what is versioned for rollback.
- Inspect the scoring and ranking constraints reference and identify supported signals.
- Review evaluation methodology: offline reports, metrics, and how to gate changes.
- Find where experimentation hooks and attribution/exposure logging are documented.

## Go-to pages

- **[How it works](../explanation/how-it-works.md)**
- **[Scoring model specification](../recsys-algo/scoring-model.md)**
- **[Ranking & constraints reference](../recsys-algo/ranking-reference.md)**
- **[Evaluation checklist](../recsys-engineering/evaluation-checklist.md)**
- **[Offline gate in CI](../recsys-eval/docs/workflows/offline-gate-in-ci.md)**

## Read next

- [Exposure logging & attribution](../explanation/exposure-logging-and-attribution.md)
- [Experimentation model](../explanation/experimentation-model.md)
- [Metrics reference](../recsys-eval/docs/metrics.md)
- [Pipelines operational invariants](../explanation/pipelines-operational-invariants.md)
