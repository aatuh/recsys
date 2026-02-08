---
diataxis: how-to
tags:
  - how-to
  - integration
  - developer
  - recsys-service
---
# Integration cookbooks (map RecSys to your domain)
This guide shows how to integration cookbooks (map RecSys to your domain) in a reliable, repeatable way.


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

- Basic integration steps: [How-to: integrate recsys-service into an application](../integrate-recsys-service.md)
- Exposure logging & attribution: [Exposure logging and attribution](../../explanation/exposure-logging-and-attribution.md)
- Event join logic: [Event join logic (exposures ↔ outcomes ↔ assignments)](../../reference/data-contracts/join-logic.md)

## Cookbooks

- Webshop: [Cookbook: integrate with a webshop](webshop.md)
- Content feed: [Cookbook: integrate with a content feed](content-feed.md)
- Event bus / streaming: [Cookbook: integrate with an event bus (streaming)](event-bus.md)

## Read next

- Back to How-to guides: [How-to guides](../index.md)
- Troubleshoot integration failures: [How-to: troubleshooting for integrators](../troubleshooting-integration.md)
- Integration contract: [Integration spec](../../reference/integration-spec.md)
