---
tags:
  - explanation
  - overview
  - business
  - developer
  - ml
---

# Capability matrix (scope and non-scope)

## Who this is for

- Evaluators who need a quick “fit” check (what RecSys is built to do).
- Recommendation engineers who want to understand what is deterministic vs data-dependent.

## What you will get

- A 2-minute scope check: what is supported vs intentionally out of scope.
- Links to the canonical pages for each capability.

## The matrix (current)

| Area | What’s included (this repo) | What’s intentionally not included (by default) | Where to read more |
| --- | --- | --- | --- |
| Serving API | `POST /v1/recommend`, tenancy/auth, limits, caching | Managed hosting | API reference: [`reference/api/api-reference.md`](../reference/api/api-reference.md) |
| Determinism | Deterministic ranking for the same inputs + versions | KPI lift guarantees | Determinism contract: [`explanation/how-it-works.md`](how-it-works.md) |
| Ranking control | Rules (pin/exclude), constraints, stable ordering | “Black-box” end-to-end models in the serving stack | Ranking reference: [`recsys-algo/ranking-reference.md`](../recsys-algo/ranking-reference.md) |
| Data modes | DB-only start + artifact/manifest mode for versioned ship/rollback | Implicit “auto-sync” of manifests without an explicit publish step | Data modes: [`explanation/data-modes.md`](data-modes.md) |
| Audit trail | Exposure logging + join by `request_id` | Logging raw PII as a requirement | Attribution: [`explanation/exposure-logging-and-attribution.md`](exposure-logging-and-attribution.md) |
| Evaluation | Offline and online evaluation workflows; ship/hold/rollback decisions | “One metric to rule them all” defaults for every domain | Workflow: [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md) |
| Operations | Runbooks, failure modes, readiness checklist | “Set-and-forget” operations | Ops hub: [`operations/index.md`](../operations/index.md) |
| Multi-tenancy | Tenant-scoped config, rules, and data isolation | Auto-provisioning tenants via an admin create-tenant API | Known limitations: [`start-here/known-limitations.md`](../start-here/known-limitations.md) |

## Notes

- The suite is designed for **operational predictability first**: deterministic serving, clear audit artifacts, and
  explicit rollback levers.
- If you need a single “no-surprises” list of limitations and non-goals, start here:
  [`start-here/known-limitations.md`](../start-here/known-limitations.md)

## Next steps

- Stakeholder overview: [`start-here/what-is-recsys.md`](../start-here/what-is-recsys.md)
- Architecture and data flow: [`explanation/how-it-works.md`](how-it-works.md)
- Known limitations: [`start-here/known-limitations.md`](../start-here/known-limitations.md)
