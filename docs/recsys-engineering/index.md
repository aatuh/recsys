---
diataxis: explanation
tags:
  - start-here
  - recsys-engineering
  - ml
  - evaluation
---
# RecSys engineering hub
Use this hub to understand ranking behavior, signals, and how to validate changes end-to-end.


## Who this is for

- Recommendation engineers (RecEng / ML engineers / data scientists)
- Anyone who needs to understand **ranking behavior**, **signals**, and **how to validate changes**.

## What you will get

- A fast path to “I can reason about the ranking output”
- A map of what you can change **without code**, **with pipelines**, and **with ranking code**
- A practical evaluation workflow (offline gate + online validation)

## 10-minute path

1. **Mental model (how data turns into ranked items)**  
   Read: [How it works: architecture and data flow](../explanation/how-it-works.md)

2. **What the ranking core does and what’s deterministic**  
   Read: [Ranking & constraints reference](../recsys-algo/ranking-reference.md)

3. **How to decide “ship / hold / rollback” (with a written trail)**  
   Read: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)


4. **Validate measurement loop (logging + joinability)**
   Read: [How-to: validate logging and joinability](../how-to/validate-logging.md)

5. **Evaluate a change**: run an offline report and review expected metrics and tradeoffs — start at [Evaluation hub](../evaluate/index.md).
   - For CI gating, see [Offline gate in CI](../recsys-eval/docs/workflows/offline-gate-in-ci.md).

## The knobs you can turn

### Without code changes (fast iteration)

- **Weights / limits / flags per tenant** (admin config)  
  Reference: [Admin API + local bootstrap (recsys-service)](../reference/api/admin.md)

- **Merchandising rules** (pin, block, boosts, constraints)  
  How-to: [How-to: integrate recsys-service into an application](../how-to/integrate-recsys-service.md)

- **Data mode choice** (DB-only vs artifact/manifest)  
  Explanation: [Data modes: DB-only vs artifact/manifest](../explanation/data-modes.md)

### With pipeline changes (data changes, stable serving)

- **New or improved signals** (popularity / co-occurrence / embeddings, etc.)  
  How-to: [How-to: add a new signal end-to-end](../how-to/add-signal-end-to-end.md)

- **Artifact + manifest lifecycle** (publish, rollback, freshness)  
  Explanation: [Artifacts and manifest lifecycle (pipelines → service)](../explanation/artifacts-and-manifest-lifecycle.md)

### With ranking code changes (high leverage, requires evaluation)

- **Candidate merge, scoring, tie-break rules**  
  Reference: [Scoring model specification (recsys-algo)](../recsys-algo/scoring-model.md)  
  Reference: [Ranking & constraints reference](../recsys-algo/ranking-reference.md)

## Evaluation workflow (practical)

1. **Make a change** (config/rules/signal/ranking)
2. **Run offline evaluation gates** (deterministic pass/fail)  
   See: [CI gates: using recsys-eval in automation](../recsys-eval/docs/ci_gates.md)
3. **Interpret the results** (metrics + tradeoffs)  
   See: [Interpreting results: how to go from report to decision](../recsys-eval/docs/interpreting_results.md)  
   Orientation: [Interpreting metrics and reports](../explanation/metric-interpretation.md)
4. **Decide ship/hold/rollback**  
   See: [Decision playbook: ship / hold / rollback](../recsys-eval/docs/decision-playbook.md)

!!! tip "If you're new to evaluation"
    Start with the suite workflow: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)

## Concepts worth reading (when you have 30 minutes)

- Evaluation validity: what numbers mean, and what they *don’t*: [Evaluation validity](../explanation/eval-validity.md)
- Guarantees and non-goals (blunt): [Guarantees and non-goals](../explanation/guarantees-and-non-goals.md)
- Ethics and fairness notes: [Ethics and fairness notes](../explanation/ethics-and-fairness.md)

## Read next

- Customization map: [Customization map](../explanation/customization-map.md)
- Verify determinism: [Verify determinism](../tutorials/verify-determinism.md)
- Verify joinability: [Verify joinability (request IDs → outcomes)](../tutorials/verify-joinability.md)
- Tune ranking safely: [How-to: tune ranking safely](../how-to/tune-ranking.md)
