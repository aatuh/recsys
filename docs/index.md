---
diataxis: explanation
tags:
  - start-here
  - overview
  - business
  - developer
  - ops
  - ml
---
# RecSys documentation

RecSys is a modular recommendation suite you can run locally or in production. The docs are organized by **what you want to do** and by **persona entry points** so you can reach the next needed page in ≤2 clicks.

## Choose your path (persona hubs)

- **Lead Developer (integration + deployment)** → [Developers hub](developers/index.md)
- **Business representative (evaluation + procurement)** → [For businesses](for-businesses/index.md)
- **SRE / operator (reliability + runbooks)** → [Operations](operations/index.md)
- **Recommendation engineer (signals + evaluation)** → [RecSys engineering hub](recsys-engineering/index.md)
- **Technical writer / docs lead (quality + consistency)** → [Technical writer persona](personas/technical-writer.md)

## 10-minute outcomes

Pick one goal and follow the linked page end-to-end.

- Get a ranked list from the API: [Tutorial: Quickstart (minimal)](tutorials/quickstart-minimal.md)
- Run a credible 1-surface pilot: [How-to: minimum pilot setup (one surface)](how-to/pilot-minimum-setup.md)
- Understand the architecture and data flow: [How it works: architecture and data flow](explanation/how-it-works.md)

## What you get (in plain terms)

- A **serving API** (`recsys-service`) you call from your product to get ranked items.
- A **pipeline** (`recsys-pipelines`) that builds signals/artifacts and publishes a manifest for safe shipping and rollback.
- An **evaluation toolchain** (`recsys-eval`) to validate data joins and compute decision-ready metrics.

## Buying / adoption in ≤3 clicks

If you are evaluating RecSys for your organization:

1. Start here: [Business representative hub](for-businesses/index.md)
2. Understand licensing: [Licensing](licensing/index.md)
3. Decide and proceed: [Pricing and ordering](pricing/index.md)

Procurement artifacts:

- [Procurement pack](for-businesses/procurement-pack.md)

## Evidence, benchmarks, and limitations

- What performance you should expect: [Benchmarks (buyer-facing)](for-businesses/benchmarks.md)
- How to reproduce baseline numbers: [Baseline benchmarks (ops)](operations/baseline-benchmarks.md)
- What RecSys does *not* promise: [Guarantees and non-goals](explanation/guarantees-and-non-goals.md)

## Security and privacy posture

- Overview and checklist: [Security, privacy, compliance](start-here/security-privacy-compliance.md)

## If you are lost

- Map of the docs: [Docs map](start-here/docs-map.md)
- Suite reference entry point: [Reference](reference/index.md)
