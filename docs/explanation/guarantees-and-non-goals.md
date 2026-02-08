---
diataxis: explanation
tags:
  - explanation
  - scope
  - trust
---
# Guarantees and non-goals

This suite is designed to be **auditable**, **deterministic**, and **operationally safe**. It is not designed to magically produce business lift without the usual work: instrumentation, iteration, and product judgment.

## Guarantees (what we claim)

### Deterministic ranking outputs (given fixed inputs)

- With the same inputs and configuration, ranking output is deterministic.
- Artifact/manifest mode provides a stable “what version served this request?” story.

See: [Verify determinism](../tutorials/verify-determinism.md) and [Artifacts and manifest lifecycle (pipelines → service)](artifacts-and-manifest-lifecycle.md).

### Clear contracts for attribution and joins

- `request_id` ties together response → exposure → outcome.
- Join logic is explicit and documented.

See: [Event join logic (exposures ↔ outcomes ↔ assignments)](../reference/data-contracts/join-logic.md) and [Integration spec (one surface)](../reference/integration-spec.md).

### Safe ship / hold / rollback loop

- A change can be evaluated, shipped, and rolled back with a clear decision trail.

See: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md).

### Operational visibility

- Health endpoints, logs, and runbooks exist to diagnose common failures.

See: [Operations](../operations/index.md) and [Runbook: Service not ready](../operations/runbooks/service-not-ready.md).

## Non-goals (what we explicitly do not claim)

### “We guarantee revenue lift”

No recommender can guarantee lift without controlling for data quality, product changes, seasonality, and measurement error.

RecSys provides tooling to *measure* and to *decide*.

### “One model fits all”

RecSys ships with practical default behaviors and a customization map, but it does not assume a single best ranking approach across all domains.

See: [Customization map](customization-map.md).

### “End-to-end user identity resolution”

The suite assumes you can provide stable pseudonymous user/session identifiers and a join key (`request_id`). Identity graphs and cross-device resolution are out of scope.

### “A full feature store / MLOps platform”

Pipelines produce artifacts and a manifest pointer. RecSys does not attempt to replace your data platform.

See: [recsys-pipelines docs](../recsys-pipelines/docs/index.md).

## Practical interpretation

If you need guarantees beyond the list above, treat that as a red flag. Capture it explicitly during evaluation and document it as either:

- an integration requirement,
- an operational requirement,
- or a product requirement (non-technical).

## Read next

- Evaluation validity: [Evaluation validity](eval-validity.md)
- Known limitations: [Known limitations and non-goals (current)](../start-here/known-limitations.md)
- Security, privacy, compliance: [Security, privacy, and compliance (overview)](../start-here/security-privacy-compliance.md)
