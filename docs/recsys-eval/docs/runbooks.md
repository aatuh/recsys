# Runbooks: operating recsys-eval

## Who this is for

Maintainers and on-call engineers.

## What you will get

- The top failure modes and how to debug them quickly
- A repeatable "triage" flow

## Triage flow

1. Identify the run:

   - run_id
   - mode
   - dataset window
   - binary version

1. Check data quality:

   - schema validation
   - duplicates
   - missing required fields

1. Check joins:

   - match rates
   - timestamp anomalies

1. Check gates and warnings:

   - which metric triggered the gate
   - which segment drove the regression

1. Decide action:

   - fix data
   - rerun
   - rollback config/model
   - escalate

## Failure mode: schema validation fails

Symptoms:

- validate command reports missing fields or wrong types

Fix:

- update logging to match schemas
- if schema changed, bump schema version and update producers

## Failure mode: join match rate collapses

Symptoms:

- offline metrics drop to near zero
- report shows low join match

Likely causes:

- request_id changed format
- producers stopped logging outcomes with request_id
- duplicate or missing request_id in exposures

Fix:

- compare recent exposure and outcome samples
- confirm request_id consistency end-to-end

## Failure mode: SRM warning (experiments)

Symptoms:

- control vs candidate sample sizes are off

Likely causes:

- bucket assignment bug
- logging bug
- rollout was not actually 50/50

Fix:

- stop interpreting metrics
- fix assignment and rerun

## Failure mode: OPE high variance

Symptoms:

- warnings about near-zero propensities
- wildly unstable estimates

Fix:

- do not ship based on OPE
- improve propensity logging and overlap
- prefer A/B or interleaving

## Read next

- Troubleshooting: [`recsys-eval/docs/troubleshooting.md`](troubleshooting.md)
- Online A/B workflow: [`recsys-eval/docs/workflows/online-ab-in-production.md`](workflows/online-ab-in-production.md)
- CI gates: [`recsys-eval/docs/ci_gates.md`](ci_gates.md)
- Metrics: [`recsys-eval/docs/metrics.md`](metrics.md)
