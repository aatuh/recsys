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

--8<-- "_snippets/key-terms.list.snippet"
--8<-- "_snippets/key-terms.defs.one-up.snippet"

## Quick paths

<div class="grid cards" markdown>

- **[Quickstart (10 minutes)](../tutorials/quickstart.md)**  
  Fastest path to a non-empty `POST /v1/recommend` + an exposure log.
- **[Run end-to-end locally](../tutorials/local-end-to-end.md)**  
  20–30 min tutorial to run the full loop on your laptop.
- **[Choose your data mode](../start-here/choose-data-mode.md)**  
  Decide DB-only vs artifact/manifest mode and jump to the right tutorial.
- **[First surface end-to-end](../how-to/first-surface-end-to-end.md)**  
  One stitched path: choose mode → integrate API → instrument → run first eval.
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
- Integration spec (headers, tenant scope, request_id): [`developers/integration-spec.md`](integration-spec.md)

Quick sanity check:

- [ ] `request_id` is stable across exposure + outcome logs
- [ ] join-rate is not near-zero
- [ ] at least one KPI and one guardrail metric exists

## Need guided help?

- RecSys Copilot (Custom GPT):
  [`chatgpt.com/g/.../recsys-copilot`](https://chatgpt.com/g/g-68c82a5c7704819185d0ff929b6fff11-recsys-copilot)

## Read next

- Exposure logging & attribution:
  [Exposure logging & attribution](../explanation/exposure-logging-and-attribution.md)
- Candidate vs ranking: [Candidate vs ranking](../explanation/candidate-vs-ranking.md)
- Ranking reference (signals, knobs, determinism): [Ranking reference](../recsys-algo/ranking-reference.md)
- Security, privacy, compliance:
  [Security/privacy/compliance](../start-here/security-privacy-compliance.md)
