---
diataxis: explanation
tags:
  - explanation
  - evaluation
  - recsys-eval
  - trust
---
# Evaluation validity

Evaluation numbers are only useful when you can explain **what they mean**, **what they do not mean**, and **what you would do if they move**.

This page makes the evaluation claims in this documentation precise and conservative.

## What RecSys guarantees about evaluation

RecSys can guarantee **process quality** (repeatability and auditability), not business outcomes.

- **Reproducible offline runs**: given the same inputs and the same artifact/manifest versions, you can re-run offline evaluation and obtain the same outputs.
- **Auditable joins**: exposures, outcomes, and assignments are joined by `request_id` with explicit join logic.
- **Decision trail**: reports are intended to be stored alongside the change that produced them, so “why did we ship?” is answerable.

See also: [Guarantees and non-goals](guarantees-and-non-goals.md).

## What RecSys does not guarantee

- That offline uplift translates to online uplift.
- That an online experiment result generalizes beyond its traffic slice.
- That any particular metric is “the metric.”

RecSys helps you *measure* and *decide*; it does not remove the need for product judgment.

## The 4 failure modes of evaluation

### 1) Unjoinable logs

Symptoms:

- Join-rate is near zero.
- Metrics look “too good” or “too empty.”

Fix:

- Implement the attribution contract: [Integration spec (one surface)](../reference/integration-spec.md)
- Verify joinability: [Verify joinability (request IDs → outcomes)](../tutorials/verify-joinability.md)

### 2) Leaky or biased offline datasets

Common causes:

- Training/eval data includes signals that encode outcomes (data leakage).
- Offline dataset does not represent production traffic.
- The candidate set differs from what serving would produce.

Mitigations:

- Treat offline evaluation as a **gate** (reject obvious regressions), not as the final word.
- Keep an “offline validity checklist” in your evaluation runbook.

### 3) Metric mismatch

Common causes:

- Optimizing proxy metrics that do not align with business goals.
- Ignoring guardrails (diversity, latency, bounce, refunds, etc.).

Mitigations:

- Define **one KPI + one guardrail** per surface first. See: [Success metrics (KPIs, guardrails, and exit criteria)](../for-businesses/success-metrics.md)
- Use a decision rubric (ship/hold/rollback). See: [Decision playbook: ship / hold / rollback](../recsys-eval/docs/decision-playbook.md)

### 4) Overconfidence from small samples

Common causes:

- Shipping based on tiny traffic slices.
- Multiple comparisons without discipline.

Mitigations:

- Prefer conservative thresholds.
- Use interleaving or OPE only when you can justify assumptions.

References:

- OPE: [Off-policy evaluation (OPE): powerful and easy to misuse](../recsys-eval/docs/ope.md)
- Interleaving: [Interleaving: fast ranker comparison on the same traffic](../recsys-eval/docs/interleaving.md)

## Recommended “trustable” evaluation ladder

Use this ladder to move from “safe to merge” to “safe to ship”:

1. **Determinism + invariants** (CI gate)
   - Verify determinism: [Verify determinism](../tutorials/verify-determinism.md)
   - Pipelines invariants: [Pipelines operational invariants (safety model)](pipelines-operational-invariants.md)

2. **Offline gate** (reject obvious regressions)
   - Workflow: [Workflow: Offline gate in CI](../recsys-eval/docs/workflows/offline-gate-in-ci.md)

3. **Online validation** (confirm impact in production)
   - Workflow: [Workflow: Online A/B analysis in production](../recsys-eval/docs/workflows/online-ab-in-production.md)

4. **Decision + documentation**
   - Suite how-to: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)

## Read next

- Guarantees and non-goals: [Guarantees and non-goals](guarantees-and-non-goals.md)
- Evaluation modes: [Evaluation modes](evaluation-modes.md)
- Exposure logging & attribution: [Exposure logging and attribution](exposure-logging-and-attribution.md)
