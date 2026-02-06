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

## Evidence ladder (how to interpret)

This page shows example artifacts you can generate in a credible pilot.

What this evidence **does** prove:

- You can serve non-empty recommendations (`POST /v1/recommend`).
- You can log what was shown (exposures) and what happened (outcomes), and join them by `request_id`.
- You can produce a shareable report and make a ship/hold/rollback decision with an audit trail.

What this evidence **does not** prove by itself:

- KPI lift in your product (you still need your own data + experimentation discipline).
- Production readiness (use the production checklist + runbooks):
  [Production readiness checklist](../operations/production-readiness-checklist.md)
- Absolute performance/latency guarantees (use baseline anchor numbers as a starting point):
  [Baseline benchmarks](../operations/baseline-benchmarks.md)

## The artifacts that matter

### 1) Serving output (what users see)

An API response includes ranked items plus metadata and warnings.

Example (response shape, abbreviated):

```json
{
  "items": [{ "item_id": "item_3", "rank": 1, "score": 0.12 }],
  "meta": {
    "tenant_id": "demo",
    "surface": "home",
    "config_version": "W/\"...\"",
    "rules_version": "W/\"...\"",
    "request_id": "req-1"
  },
  "warnings": []
}
```

Start here:

- Local end-to-end tutorial: [Local end-to-end](../tutorials/local-end-to-end.md)

### 2) Exposure/outcome logs (what we measure)

You need auditable logs to attribute outcomes to what was shown.

Examples (JSONL; shown pretty-printed for readability):

Exposure (`exposure.v1`):

```json
{
  "request_id": "req-1",
  "user_id": "u_1",
  "ts": "2026-02-05T10:00:00Z",
  "items": [
    { "item_id": "item_1", "rank": 1 },
    { "item_id": "item_2", "rank": 2 }
  ],
  "context": { "tenant_id": "demo", "surface": "home" }
}
```

Outcome (`outcome.v1`):

```json
{
  "request_id": "req-1",
  "user_id": "u_1",
  "item_id": "item_2",
  "event_type": "click",
  "ts": "2026-02-05T10:00:02Z"
}
```

Start here:

- Data contracts (schemas + examples): [Data contracts](../reference/data-contracts/index.md)
- Exposure logging and attribution:
  [Exposure logging and attribution](../explanation/exposure-logging-and-attribution.md)

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

- Suite workflow: [Run eval and ship](../how-to/run-eval-and-ship.md)
- recsys-eval overview: [recsys-eval overview](../recsys-eval/overview.md)

### 4) Audit record (what changed, and who changed it)

Control-plane changes (config/rules/cache invalidation) can be written to an audit log.

Example (abbreviated):

```json
{
  "tenant_id": "demo",
  "entries": [
    {
      "id": 123,
      "occurred_at": "2026-02-05T10:00:00Z",
      "actor_sub": "user:demo-admin",
      "actor_type": "user",
      "action": "config.update",
      "entity_type": "tenant_config",
      "entity_id": "demo",
      "request_id": "req-1"
    }
  ]
}
```

Start here:

- Admin API bootstrap and audit endpoint: [Admin API](../reference/api/admin.md)
- Security hardening checklist (includes audit logging):
  [Security, privacy, compliance](../start-here/security-privacy-compliance.md)

## A reproducible demo path (under an hour)

Run the suite locally and produce a report you can share:

- Tutorial: [Local end-to-end](../tutorials/local-end-to-end.md)

This gives you:

- a working serving API
- eval-compatible exposure logs
- a minimal outcome log
- a sample evaluation report

## Read next

- Success metrics (KPIs + guardrails): [Success metrics](success-metrics.md)
- Evaluation, pricing, and licensing (buyer guide): [Buyer guide](../pricing/evaluation-and-licensing.md)
- Security pack: [Security pack](../security/security-pack.md)
