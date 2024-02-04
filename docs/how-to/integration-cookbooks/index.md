# Integration cookbooks (map RecSys to your domain)

## Who this is for

Integrators and platform engineers who need a quick, concrete mapping from RecSys concepts to real product patterns.

## What you will get

- Three “copy this mental model” integration patterns
- What to log (and why) so evaluation and troubleshooting work
- Pitfalls that commonly break attribution and trust

## Common building blocks (all cookbooks)

- Define stable surfaces (`home`, `pdp`, `related`, etc.).
- Generate or propagate a stable `request_id` per rendered list.
- Log exposures (ranked list) and outcomes (click/conversion) with the same `request_id`.

Read first if you haven’t yet:

- Basic integration steps: [`how-to/integrate-recsys-service.md`](../integrate-recsys-service.md)
- Exposure logging & attribution: [`explanation/exposure-logging-and-attribution.md`](../../explanation/exposure-logging-and-attribution.md)
- Event join logic: [`reference/data-contracts/join-logic.md`](../../reference/data-contracts/join-logic.md)

## Cookbooks

- Webshop: [`webshop.md`](webshop.md)
- Content feed: [`content-feed.md`](content-feed.md)
- Event bus / streaming: [`event-bus.md`](event-bus.md)
