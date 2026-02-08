---
diataxis: explanation
tags:
  - business
  - executive-summary
  - procurement
---
# Executive summary

RecSys is a production-ready recommendation system suite designed for **predictable operations**: deterministic serving,
auditable logs, and safe ship/rollback.

## Outcomes you can expect

- **Faster iteration** on ranking behavior without “black box” surprises
- **Lower operational risk** through deterministic serving and explicit rollback paths
- **Measurable impact** via offline evaluation gates and online experimentation support

## What you get

- **Serving API** for online recommendations (tenant-scoped, surface-aware)
- **Ranking core** with explainable scoring and deterministic behavior
- **Pipelines** for building versioned artifacts and managing freshness (optional)
- **Evaluation** for offline metrics and decision support (ship/hold/rollback)

See the full scope: [Procurement pack (Security, Legal, IT, Finance)](procurement-pack.md)

## What you need to provide

- Your **items** and **surfaces** (home, PDP, cart, …)
- A way to emit **exposure logs** (what was shown) and ideally **outcome logs** (what happened)
- Your integration choice: DB-only pilot vs artifact/manifest (production-like)

Start here: [Minimum components by goal](../start-here/minimum-components-by-goal.md)

## What RecSys is optimized for

- Teams that want **control** (rules + weights + deterministic behavior)
- Teams that need **auditability** and evaluation-ready data
- Teams that prefer **incremental adoption** (pilot in DB-only mode, then scale)

## Fast next steps (no meetings required)

1. Run the **10-minute Quickstart**: [Quickstart (10 minutes)](../tutorials/quickstart.md)
2. Read the **buyer guide**: [Evaluation, pricing, and licensing (buyer guide)](../pricing/evaluation-and-licensing.md)
3. Follow the **pilot plan**: [Pilot plan (2–6 weeks)](../start-here/pilot-plan.md)

## Read next

- Buyer journey: [Buyer journey: evaluate RecSys in 5 minutes](buyer-journey.md)
- Procurement pack: [Procurement pack (Security, Legal, IT, Finance)](procurement-pack.md)
- Evaluation guide: [Start an evaluation](../evaluate/index.md)
