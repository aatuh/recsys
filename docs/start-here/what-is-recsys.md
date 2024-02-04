# What the RecSys suite is (stakeholder overview)

## Who this is for

Product managers, business stakeholders, and engineering leads evaluating whether to adopt the RecSys suite.

## What you will get

- A plain-language description of what the suite does (and does not do)
- Typical use cases and success metrics
- The minimum data you must provide
- A realistic pilot → production timeline

## What the RecSys suite is

The **RecSys suite** is an end-to-end, production-oriented recommendation system toolkit:

- `recsys-service`: the online HTTP API (auth, tenancy, caching, limits, exposure logging)
- `recsys-algo`: the deterministic ranking core (candidate merge, scoring, constraints, rules)
- `recsys-pipelines`: offline pipelines that turn events into versioned signals/artifacts
- `recsys-eval`: evaluation tooling to measure quality and decide what to ship

Design intent: **predictable, auditable behavior** before “black-box” modeling. You can roll out changes safely, explain
what changed, and roll back by version.

## What you can build with it

Common product placements:

- Homepage / feed recommendations (“for you”, “popular now”)
- PDP “similar items”
- “Related content” blocks (articles, videos, collections)

Common operational needs it supports:

- Controlled outcomes via rules (pin/exclude, constraints)
- Multi-tenant operation (per-org configuration and isolation)
- “Ship/rollback” discipline (versioned config/rules/manifests)

## How success is measured

You typically measure success in two layers:

1) **Business metrics** (what leadership cares about)
   - click-through rate (CTR), conversion rate (CVR), revenue per session
   - retention / repeat visits (when applicable)

2) **Recommendation metrics** (what helps you iterate safely)
   - offline metrics like hitrate@k and precision@k (in `recsys-eval`)
   - online experiments (A/B) to validate impact in production

Operationally, you also track:

- P95/P99 latency and error rate
- “empty recs” rate (a key indicator of data/config issues)

## What data you need (minimum)

You need identifiers and events that can connect “what was served” to “what the user did”.

Minimum inputs:

- An item catalog (`item_id` + basic metadata/tags)
- Interaction events (click/purchase/etc) with:
  - `tenant_id` (org/account)
  - `item_id`
  - `event_type`
  - timestamp (`occurred_at`)

Strongly recommended for evaluation:

- A stable, pseudonymous `user_id` (do not log raw PII)
- A `request_id` you can propagate from serving → outcomes so evaluation can join correctly
- Exposure logs (the ranked list you showed, with ranks)

See: [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)

## Timeline: pilot → production (typical)

Every company’s data readiness differs, but a realistic plan is:

- **Week 1**: baseline + instrumentation
  - stand up the service in a dev environment (DB-only mode is the fastest start)
  - integrate a placement and log exposures + outcomes
  - run `recsys-eval` on real logs to validate joins and sanity metrics

- **Weeks 2–3**: improve relevance safely
  - add similarity/co-visitation signals via pipelines (optional but high ROI)
  - introduce simple rules (pin/exclude) for business control
  - iterate with offline evaluation, then ship to a small audience

- **Weeks 4–6**: production hardening + experimentation
  - set SLOs and add runbooks to on-call rotation
  - enable controlled A/B tests for key surfaces
  - expand to more placements and segments

## Read next

- Start here (engineers): [`start-here/index.md`](index.md)
- Pilot plan: [`start-here/pilot-plan.md`](pilot-plan.md)
- Local end-to-end tutorial: [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)
- Known limitations: [`start-here/known-limitations.md`](known-limitations.md)
