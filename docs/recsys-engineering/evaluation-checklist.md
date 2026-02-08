---
diataxis: explanation
tags:
  - recsys-engineering
  - evaluation
  - checklist
  - ml
---
# Evaluation checklist (recommendation engineer)
This page explains Evaluation checklist (recommendation engineer) and how it fits into the RecSys suite.


## Who this is for

- Recommendation engineers validating changes to signals, candidates, or ranking behavior
- Anyone responsible for a “ship/hold/rollback” decision backed by evidence

## What this is

A practical checklist that catches the most common evaluation failures:

- bad joins (request_id problems)
- leakage (training/eval contamination)
- non-comparable baselines
- metrics that look good but don’t translate to user impact

This is not a textbook. It is a “don’t ship blind” list.

## 0) Define the decision and scope

- [ ] What is the decision? ship / hold / rollback
- [ ] What changed? config / rules / signal / ranking code
- [ ] What is the target surface? `surface = ...`
- [ ] What is the primary KPI and minimum effect size?
- [ ] What guardrails must not regress?

See: [Success metrics (KPIs, guardrails, and exit criteria)](../for-businesses/success-metrics.md)

## 1) Data and join sanity (mandatory)

- [ ] `request_id` is stable per render and present in:
  - [ ] serving response `meta.request_id`
  - [ ] exposure events
  - [ ] outcome events
- [ ] Join-rate is measured and acceptable for the surface
- [ ] Exposure ranks are recorded (position bias matters)
- [ ] Identifiers are pseudonymous (avoid raw PII)

Canonical specs:

- Minimum instrumentation: [Minimum instrumentation spec (for credible evaluation)](../reference/minimum-instrumentation.md)
- Join logic: [Event join logic (exposures ↔ outcomes ↔ assignments)](../reference/data-contracts/join-logic.md)

## 2) Baseline comparability (mandatory)

- [ ] Baseline is clearly defined (what system/logic, what parameters)
- [ ] Same population, same time window, same filters
- [ ] You can reproduce baseline numbers

See: [Baseline benchmarks (anchor numbers)](../operations/baseline-benchmarks.md)

## 3) Offline evaluation gate (recommended before any online test)

- [ ] Run offline evaluation with a deterministic snapshot (or clearly define sampling)
- [ ] Ensure no leakage (training data leaking into evaluation window)
- [ ] Inspect both aggregate metrics and slices (new users, cold-start, long-tail)

Start here:

- Suite workflow: [How-to: run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)
- Offline gate in CI: [Workflow: Offline gate in CI](../recsys-eval/docs/workflows/offline-gate-in-ci.md)
- Interpreting results: [Interpreting results: how to go from report to decision](../recsys-eval/docs/interpreting_results.md)

## 4) Online validation (when you have traffic)

- [ ] Choose test type: A/B, interleaving, or staged rollout
- [ ] Confirm randomization and bucketing are stable
- [ ] Track guardrails in near real-time (latency, empty-recs rate, errors)
- [ ] Predefine stop conditions (kill switch thresholds)

See: [Workflow: Online A/B analysis in production](../recsys-eval/docs/workflows/online-ab-in-production.md)

## 5) Decision artifact (required)

- [ ] Produce a shareable decision record (what changed, what the evidence says)
- [ ] Link the report outputs and where raw logs live
- [ ] Record rollback lever and confirm it works

Evidence template: [Evidence (what “good outputs” look like)](../for-businesses/evidence.md)

## Read next

- RecSys engineering hub: [RecSys engineering hub](index.md)
- Eval validity (what numbers mean): [Evaluation validity](../explanation/eval-validity.md)
- Decision playbook: [Decision playbook: ship / hold / rollback](../recsys-eval/docs/decision-playbook.md)
