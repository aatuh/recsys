---
tags:
  - overview
  - business
---

# Use cases (where to start)

## Who this is for

- Product owners choosing the first surface to pilot
- Engineering leads scoping “smallest credible” integration work

## What you will get

- A short list of common recommendation surfaces
- What you need to log for each surface (at minimum)
- A safe “start small” recommendation per category

## E-commerce

### Home feed (“for you” / personalized shelves)

Start when:

- you can log exposures + outcomes with stable `request_id`
- you can run an A/B experiment or at least offline regression

Typical primary KPI:

- conversion rate or revenue per session

Common guardrails:

- latency, error rate, empty-recs rate

Next page(s):

- Webshop integration cookbook: [`how-to/integration-cookbooks/webshop.md`](../how-to/integration-cookbooks/webshop.md)
- Data contracts: [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)

### PDP “Similar items”

Why it’s a great pilot surface:

- clear intent signal (the viewed item is an anchor)
- easier to reason about relevance

Typical primary KPI:

- add-to-cart rate, click-through on the module

Next page(s):

- Webshop integration cookbook: [`how-to/integration-cookbooks/webshop.md`](../how-to/integration-cookbooks/webshop.md)
- Candidate vs ranking (mental model): [`explanation/candidate-vs-ranking.md`](../explanation/candidate-vs-ranking.md)

## Content / media

### “Next up” / “Continue watching/reading”

Why it’s a great pilot surface:

- users already signal intent with consumption order
- improvements often show up quickly

Typical primary KPI:

- completion rate, session depth, time spent

Next page(s):

- Content feed cookbook: [`how-to/integration-cookbooks/content-feed.md`](../how-to/integration-cookbooks/content-feed.md)

### “For you” feed (home)

Start when:

- you can segment by surface and (optionally) locale/device
- you can run at least one safe rollback drill

Typical primary KPI:

- return rate, engagement rate

Next page(s):

- Content feed cookbook: [`how-to/integration-cookbooks/content-feed.md`](../how-to/integration-cookbooks/content-feed.md)
- Operational reliability & rollback: [`start-here/operational-reliability-and-rollback.md`](../start-here/operational-reliability-and-rollback.md)

## Read next

- Success metrics and exit criteria: [`for-businesses/success-metrics.md`](success-metrics.md)
- Pilot plan (2–6 weeks): [`start-here/pilot-plan.md`](../start-here/pilot-plan.md)
- ROI and risk model: [`start-here/roi-and-risk-model.md`](../start-here/roi-and-risk-model.md)
