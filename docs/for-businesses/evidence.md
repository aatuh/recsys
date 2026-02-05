---
tags:
  - overview
  - business
  - evaluation
---

# Evidence (what “good outputs” look like)

## Who this is for

- Buyers who want proof the loop is real (not just architecture)
- Stakeholders who need to see what a decision artifact looks like

## What you will get

- The three concrete artifacts produced in a credible pilot
- Where they come from in this repo (so you can reproduce them)

## The three outputs that matter

### 1) Serving output (what users see)

An API response includes ranked items plus metadata and warnings.

Start here:

- Local end-to-end tutorial: [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)

### 2) Exposure/outcome logs (what we measure)

You need auditable logs to attribute outcomes to what was shown.

Start here:

- Data contracts (schemas + examples): [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)
- Exposure logging and attribution: [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)

### 3) Evaluation report (what you can share internally)

`recsys-eval` produces a machine-readable JSON report and optional Markdown/HTML summaries.

Example (executive summary shape, illustrative):

```json
{
  "run_id": "2026-02-05T12:34:56Z-abc123",
  "mode": "offline",
  "created_at": "2026-02-05T12:34:56Z",
  "version": "recsys-eval/vX.Y.Z",
  "summary": {
    "cases_evaluated": 12345,
    "executive": {
      "decision": "pass",
      "highlights": ["No regressions on guardrails"],
      "key_deltas": [{ "name": "primary_metric", "delta": 0.012, "relative_delta": 0.03 }]
    }
  }
}
```

Start here:

- Suite workflow: [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)
- recsys-eval overview: [`recsys-eval/overview.md`](../recsys-eval/overview.md)

## A reproducible demo path (under an hour)

Run the suite locally and produce a report you can share:

- Tutorial: [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)

This gives you:

- a working serving API
- eval-compatible exposure logs
- a minimal outcome log
- a sample evaluation report

## Read next

- Use cases (pick your first surface): [`for-businesses/use-cases.md`](use-cases.md)
- Success metrics (KPIs + guardrails): [`for-businesses/success-metrics.md`](success-metrics.md)
- Operational reliability & rollback: [`start-here/operational-reliability-and-rollback.md`](../start-here/operational-reliability-and-rollback.md)
