---
diataxis: explanation
tags:
  - explanation
  - ethics
  - fairness
  - risk
---
# Ethics and fairness notes

Recommenders change what users see. That creates responsibility.

RecSys is built for **auditability** and **control**, but it does not automatically solve ethical problems. You still need policy choices and monitoring.

## What RecSys can help with

### 1) Explainability of “why did this show?”

- Determinism + versioning gives you a stable answer to “what code/config/data produced this list?”
- Exposure logs and artifacts preserve a trail for review.

See: [Exposure logging and attribution](exposure-logging-and-attribution.md) and [Artifacts and manifest lifecycle (pipelines → service)](artifacts-and-manifest-lifecycle.md).

### 2) Guardrails and constraints

- You can express constraints (pin/block/diversity-like rules, caps, per-surface limits) via config and merchandising rules.
- You can define guardrail metrics in evaluation and require them to pass before shipping.

See: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md) and [CI gates: using recsys-eval in automation](../recsys-eval/docs/ci_gates.md).

### 3) Bias discovery (not bias elimination)

RecSys can help you detect distribution shifts and group disparities *if* you provide the necessary annotations.

Common examples:

- Category skew (over/under-recommending certain item categories)
- Supplier skew (over/under-recommending certain sellers)
- Cold-start starvation (new items never get exposure)

## What you must decide

### Sensitive attributes

If you log or use sensitive attributes (age, gender, health, etc.), you create legal and ethical obligations. Prefer avoiding these.

If you need them, document:

- lawful basis (GDPR) and data minimization reasoning
- retention and access controls
- how you will measure and mitigate disparate impact

### Feedback loops

Recommenders create feedback loops: what you show affects what you learn.

Mitigations often include:

- explicit exploration budgets
- freshness / novelty constraints
- counterfactual evaluation methods

RecSys supports controlled change workflows; it does not pick exploration strategies for you.

## Recommended baseline policy (pragmatic)

1. **Start with a conservative objective** (one KPI + one guardrail).
2. **Define “unsafe outputs”** (policy rules, disallowed categories, etc.).
3. **Require a written decision** for every shipped change.
4. **Review exposure distributions** periodically (by category/supplier/segment).

## Read next

- Guarantees and non-goals: [Guarantees and non-goals](guarantees-and-non-goals.md)
- Evaluation validity: [Evaluation validity](eval-validity.md)
- Security, privacy, compliance: [Security, privacy, and compliance (overview)](../start-here/security-privacy-compliance.md)
