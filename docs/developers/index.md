---
tags:
  - overview
  - developer
  - quickstart
---

# Developers

RecSys is an auditable recommendation system suite with deterministic ranking and versioned ship/rollback.

## Who this is for

- Lead developers / platform engineers integrating the suite
- Application engineers wiring one recommendation surface end-to-end
- Recommendation engineers validating ranking changes

## What you will get

- The fastest path to a runnable local setup
- A minimal “first integration” checklist (one surface + logging + first eval report)
- Links to the canonical API, contracts, and on-call runbooks

!!! info "Key terms"
    - **[Tenant](../project/glossary.md#tenant)**: a configuration + data isolation boundary (usually one organization).
    - **[Surface](../project/glossary.md#surface)**: where recommendations are shown (home, PDP, cart, ...).
    - **[Request ID](../project/glossary.md#request-id)**: the join key that ties together responses, exposures, and outcomes.
    - **[Exposure log](../project/glossary.md#exposure-log)**: what was shown (audit trail + evaluation input).

## Quick paths

<div class="grid cards" markdown>

- **[Quickstart (10 minutes)](../tutorials/quickstart.md)**  
  Fastest path to a non-empty `POST /v1/recommend` + an exposure log.
- **[Run end-to-end locally](../tutorials/local-end-to-end.md)**  
  20–30 min tutorial to run the full loop on your laptop.
- **[Minimum components by goal](../start-here/minimum-components-by-goal.md)**  
  Decide DB-only vs artifact/manifest mode and what you need to run.
- **[Integrate the API](../how-to/integrate-recsys-service.md)**  
  Auth, tenancy, contracts, and copy/paste examples.
- **[Integration checklist](../how-to/integration-checklist.md)**  
  One-surface checklist: identifiers, attribution, join-rate, fallbacks, rate limits.
- **[Run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)**  
  Validate logs → produce a report → decide ship/hold/rollback.
- **[API reference](../reference/api/api-reference.md)**  
  OpenAPI / Swagger UI and examples.
- **[Operations runbooks](../operations/index.md)**  
  The first pages to open when something goes wrong.

</div>

## Developer ladder (recommended)

Follow this path in order:

1. Get a non-empty recommendation response + one exposure log:
   - [`tutorials/quickstart.md`](../tutorials/quickstart.md)
2. Integrate one surface in your app (for example: `home_feed`):
   - [`how-to/integrate-recsys-service.md`](../how-to/integrate-recsys-service.md)
   - [`how-to/integration-checklist.md`](../how-to/integration-checklist.md)
3. Emit exposure logs and outcome logs with the same `request_id`:
   - [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)
   - [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)
4. Produce your first report and make a ship/hold decision:
   - [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)
5. Do one rollback drill (so you trust the lever before you need it):
   - [`start-here/operational-reliability-and-rollback.md`](../start-here/operational-reliability-and-rollback.md)

## Integration checklist

Use the canonical checklist (with anchors you can share in PRs/issues):

- [Integration checklist](../how-to/integration-checklist.md)

Quick sanity check:

- [ ] `request_id` is stable across exposure + outcome logs
- [ ] join-rate is not near-zero
- [ ] at least one KPI and one guardrail metric exists

## Need guided help?

- RecSys Copilot (Custom GPT): [`chatgpt.com/g/.../recsys-copilot`](https://chatgpt.com/g/g-68c82a5c7704819185d0ff929b6fff11-recsys-copilot)

## Read next

- Exposure logging & attribution: [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)
- Candidate vs ranking: [`explanation/candidate-vs-ranking.md`](../explanation/candidate-vs-ranking.md)
- Ranking reference (signals, knobs, determinism): [`recsys-algo/ranking-reference.md`](../recsys-algo/ranking-reference.md)
- Security, privacy, compliance: [`start-here/security-privacy-compliance.md`](../start-here/security-privacy-compliance.md)
