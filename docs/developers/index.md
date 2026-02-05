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

## Quick paths

<div class="grid cards" markdown>

- **[Run end-to-end locally](../tutorials/local-end-to-end.md)**  
  20–30 min tutorial to run the full loop on your laptop.
- **[Integrate the API](../how-to/integrate-recsys-service.md)**  
  Auth, tenancy, contracts, and copy/paste examples.
- **[Run evaluation and make ship decisions](../how-to/run-eval-and-ship.md)**  
  Validate logs → produce a report → decide ship/hold/rollback.
- **[API reference](../reference/api/api-reference.md)**  
  OpenAPI / Swagger UI and examples.
- **[Operations runbooks](../operations/index.md)**  
  The first pages to open when something goes wrong.

</div>

## Integration quickstart (one surface)

1. Run the local tutorial: [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)
2. Pick one surface (for example: `home_feed`) and integrate requests + responses:
   - [`how-to/integrate-recsys-service.md`](../how-to/integrate-recsys-service.md)
3. Emit exposure logs and outcome logs with the same `request_id`:
   - Contracts and examples: [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)
4. Produce your first report:
   - [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)
5. Do one rollback drill (so you trust the lever before you need it):
   - [`start-here/operational-reliability-and-rollback.md`](../start-here/operational-reliability-and-rollback.md)

## Evaluation-mode checklist

- [ ] `request_id` is stable across exposure + outcome logs
- [ ] Each exposure record includes `tenant_id` and `surface`
- [ ] You can compute join-rate (exposure↔outcome) and it’s not near-zero
- [ ] You have at least one KPI and one guardrail metric
- [ ] You can roll back config/rules and invalidate caches

## Need guided help?

- RecSys Copilot (Custom GPT): [`chatgpt.com/g/.../recsys-copilot`](https://chatgpt.com/g/g-68c82a5c7704819185d0ff929b6fff11-recsys-copilot)

## Read next

- Exposure logging & attribution: [`explanation/exposure-logging-and-attribution.md`](../explanation/exposure-logging-and-attribution.md)
- Candidate vs ranking: [`explanation/candidate-vs-ranking.md`](../explanation/candidate-vs-ranking.md)
- Security, privacy, compliance: [`start-here/security-privacy-compliance.md`](../start-here/security-privacy-compliance.md)
