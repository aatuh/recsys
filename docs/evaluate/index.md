---
tags:
  - overview
  - evaluation
  - business
  - developer
---

# Start an evaluation

RecSys is an auditable recommendation system suite with deterministic ranking and versioned ship/rollback.

[See pricing](../pricing/index.md){ .md-button }
[Run the open-source quickstart](../tutorials/local-end-to-end.md){ .md-button }
[Buyer guide](../pricing/evaluation-and-licensing.md){ .md-button }

[Start evaluation (commercial)][commercial_eval]{ .md-button .md-button--primary }
[Message on LinkedIn][recsys_linkedin]{ .md-button }

!!! info "Scope check (read before piloting)"
    Capability boundaries: [`explanation/capability-matrix.md`](../explanation/capability-matrix.md). Non-goals:
    [`start-here/known-limitations.md`](../start-here/known-limitations.md).

## Who this is for

- Teams who want to validate lift and operational fit before committing
- Procurement/security reviewers who need a concrete artifact trail and clear rollback story

## What you will get

- The minimum “credible pilot” instrumentation checklist
- A recommended 2–6 week plan and exit criteria
- Links to the exact docs you will run during the evaluation

--8<-- "_snippets/key-terms.list.snippet"
--8<-- "_snippets/key-terms.defs.one-up.snippet"

## What “evaluation” means here

An evaluation is successful when you can answer these questions with evidence:

- Does RecSys improve a KPI you care about (with guardrails holding)?
- Can you audit what happened (logs → reports → decision trail)?
- Can you roll back safely (config/rules/manifests) and restore a known-good state?

## Evaluation onboarding checklist (Phase 1)

- [ ] Choose one recommendation surface (for example: home feed, PDP similar-items, related content)
- [ ] Integrate `recsys-service` for that surface (auth + tenancy + request/response contract)
- [ ] Emit exposure logs and outcome logs with the same `request_id`
- [ ] Produce your first report (offline gate or experiment mode)
- [ ] Do one rollback drill (before you are on fire)

## Minimum data requirements

- Stable join key: `request_id` present in exposures and outcomes
- A pseudonymous `user_id` or `session_id` (no raw PII)
- `tenant_id` and `surface` on every exposure record

See:

- Data contracts (schemas + examples): [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)
- Exposure logging & attribution: [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)

[Start evaluation (commercial)][commercial_eval]{ .md-button .md-button--primary }
[Message on LinkedIn][recsys_linkedin]{ .md-button }

## Read next

- Pilot plan (2–6 weeks): [`start-here/pilot-plan.md`](../start-here/pilot-plan.md)
- Procurement pack (Security/Legal/IT/Finance): [`for-businesses/procurement-pack.md`](../for-businesses/procurement-pack.md)
- How-to run eval and ship decisions: [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)
- Default evaluation pack (recommended): [`recsys-eval/docs/default-evaluation-pack.md`](../recsys-eval/docs/default-evaluation-pack.md)
- Operational reliability & rollback: [`start-here/operational-reliability-and-rollback.md`](../start-here/operational-reliability-and-rollback.md)

[commercial_eval]: mailto:contact@recsys.app?subject=RecSys%20Commercial%20Evaluation
[recsys_linkedin]: https://www.linkedin.com/showcase/recsys-suite
