# Field catalog (lean)

exposure.v1:
- request_id: join key for outcomes
- served[].rank: 1-based ranking position

interaction.v1:
- request_id: recommended when available
- event_type: click/purchase/etc.

recsys-eval datasets (strict schemas):
- exposure.v1 / outcome.v1 / assignment.v1 required fields and examples:
  `reference/data-contracts/eval-events.md`
