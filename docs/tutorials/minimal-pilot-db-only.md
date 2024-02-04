# Tutorial: minimal pilot mode (DB-only, popularity baseline)

## Who this is for

Product owners and engineering leads who want a low-friction pilot of the **RecSys suite** without object storage or
offline pipelines.

## What you will get

- A popularity-only baseline you can ship safely
- The minimum instrumentation needed to measure impact
- A clear list of what this mode proves (and what it does not)

## What this pilot answers (business + engineering)

With DB-only + popularity, you can answer:

- Can we integrate a recommendation placement end-to-end (API → UI render)?
- Can we attribute outcomes to exposures (do joins work)?
- Are latency, error rate, and “empty recs” within acceptable bounds?
- Do we have the data and IDs needed to evaluate and iterate?

This pilot does not answer “how good is personalization”. It answers “is the loop operationally real”.

## Prereqs

- One surface to start (for example: `home`)
- A way to generate or propagate a stable `request_id` per rendered list
- A pseudonymous `user_id` or `session_id` you can use consistently

## Minimal data you need (DB-only signals)

Populate these Postgres tables for your tenant:

- `item_tags` (catalog + tags used for constraints and filters)
- `item_popularity_daily` (daily popularity score per item and surface/namespace)

The fastest approach is to backfill a small set of top items (for one surface) and refresh daily.

## Steps (recommended)

1. Run the runnable local tutorial once (proves the loop end-to-end):
   - [`tutorials/local-end-to-end.md`](local-end-to-end.md)

2. Replace the seed SQL with your own data loading:
   - Load `item_tags` from your catalog.
   - Load `item_popularity_daily` from page views / purchases / clicks (your choice).

3. Integrate one placement in your product:
   - Use `POST /v1/recommend` with a stable `request_id`.
   - Log exposures and outcomes with that same `request_id`.

4. Run evaluation on real logs early:
   - Validate schemas and compute join rates.

## Read next

- Pilot plan: [`start-here/pilot-plan.md`](../start-here/pilot-plan.md)
- Integrate service: [`how-to/integrate-recsys-service.md`](../how-to/integrate-recsys-service.md)
- Data contracts hub: [`reference/data-contracts/index.md`](../reference/data-contracts/index.md)
