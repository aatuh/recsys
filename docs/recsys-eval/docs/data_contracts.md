# Data contracts: what inputs look like

## Who this is for

Integrators and anyone who needs to produce valid input logs.

## What you will get

- The minimum required fields for each input type
- How the joins work
- Small example records

recsys-eval uses JSON Schemas for validation:

- schemas/exposure.v1.json
- schemas/outcome.v1.json
- schemas/assignment.v1.json
- api/schemas/report.v1.json
- api/schemas/decision.v1.json

Use the validate command before doing anything else:

```bash

recsys-eval validate --schema exposure.v1 --input exposures.jsonl
recsys-eval validate --schema outcome.v1 --input outcomes.jsonl
recsys-eval validate --schema assignment.v1 --input assignments.jsonl

```

## Record formats

### Exposure (what was shown)

Purpose:

- describes what items were recommended and in what order
- provides context for segment slicing
- acts as the "left side" of joins

Join key:

- request_id (required)

Minimal example (illustrative, not exhaustive):

```json

{
  "request_id": "req_123",
  "user_id": "u_42",
  "ts": "2026-01-27T12:00:00Z",
  "context": {
    "tenant_id": "demo",
    "surface": "home"
  },
  "items": [
    {"item_id": "A", "rank": 1},
    {"item_id": "B", "rank": 2}
  ]
}

```

Notes:

- For OPE, exposures may also include propensities. See docs/OPE.md.

### Outcome (what the user did)

Purpose:

- records the behavior you care about: click, conversion, etc.

Join key:

- request_id (required)

Minimal example:

```json

{
  "request_id": "req_123",
  "user_id": "u_42",
  "item_id": "B",
  "event_type": "click",
  "ts": "2026-01-27T12:00:05Z"
}

```

### Assignment (experiment bucket)

Purpose:

- tells which variant a request/user belongs to (control vs candidate)

Join key:

- request_id (required in this dataset contract)

Minimal example:

```json

{
  "experiment_id": "exp_home_rank_v3",
  "variant": "control",
  "request_id": "req_123",
  "user_id": "u_42",
  "ts": "2026-01-27T12:00:00Z"
}

```

## Interleaving datasets

Interleaving mode uses a different dataset config:

- ranker_a results
- ranker_b results
- outcomes (often clicks)

See configs/examples/dataset.interleaving.jsonl.yaml for the wiring.

## Join expectations and quality signals

Good joins are boring. Bad joins destroy trust.

In reports, look for:

- match rate: how many exposures have outcomes
- duplicate request_id rates
- ts anomalies
- missing segmentation context keys (for example: `tenant_id`, `surface`)

If joins look wrong, stop and fix instrumentation. Do not "tune metrics".

## Read next

- Integration logging plan: [`recsys-eval/docs/integration.md`](integration.md)
- Workflow: Offline gate in CI: [`recsys-eval/docs/workflows/offline-gate-in-ci.md`](workflows/offline-gate-in-ci.md)
- Suite-level contract index: [`reference/data-contracts/index.md`](../../reference/data-contracts/index.md)
- Troubleshooting joins: [`recsys-eval/docs/troubleshooting.md`](troubleshooting.md)
