---
diataxis: explanation
tags:
  - explanation
  - evaluation
  - experimentation
---
# Evaluation modes
This page explains Evaluation modes and how it fits into the RecSys suite.


## Who this is for

- Stakeholders and engineers deciding how to validate improvements
- Teams setting up a “ship / hold / rollback” workflow

## What you will get

- A map of offline vs online evaluation (what each can prove)
- How RecSys supports each mode (and what you must provide)

## Offline evaluation (deterministic, repeatable)

Goal:

- Answer: “Is this change likely better, and did we break anything?”

Requires:

- Exposure logs (what was shown)
- Outcome logs (what happened), ideally
- A defined evaluation dataset window

Docs:

- Overview and workflows: [recsys-eval docs](../recsys-eval/docs/index.md)
- CI gates: [CI gates: using recsys-eval in automation](../recsys-eval/docs/ci_gates.md)

When to use:

- Every change that affects ranking behavior (rules, weights, signals, scoring)
- As a deterministic “quality gate” before production rollout

## Online evaluation (experiments, production validation)

Goal:

- Answer: “Does this change improve business metrics under real traffic?”

Requires:

- A way to assign traffic to variants
- Stable subject IDs (user/session) to avoid broken bucketing
- Joinable logs (exposures + outcomes)

Docs:

- Experimentation model: [Experimentation model (A/B, interleaving, OPE)](experimentation-model.md)
- Online A/B workflow: [Workflow: Online A/B analysis in production](../recsys-eval/docs/workflows/online-ab-in-production.md)

When to use:

- After offline gates pass
- When you need a procurement-grade proof of impact (business KPIs)

## Interleaving (faster online comparison)

Interleaving compares rankers by mixing results in the same list.

Docs:

- Interleaving: [Interleaving: fast ranker comparison on the same traffic](../recsys-eval/docs/interleaving.md)

## Read next

- Run eval and ship: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)
- Interpretation cheat sheet: [Interpretation cheat sheet (recsys-eval)](../recsys-eval/docs/workflows/interpretation-cheat-sheet.md)
- Experimentation model: [Experimentation model (A/B, interleaving, OPE)](experimentation-model.md)
