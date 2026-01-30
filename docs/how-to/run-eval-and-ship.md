# How-to: run evaluation and make ship decisions

Always run an offline regression gate:
- compare baseline vs candidate versions
- fail if primary metric regresses beyond threshold

Prefer online A/B tests:
- log exposures with experiment id/variant
- log outcomes tied to the same request_id
- check KPI + guardrails

Ship if KPI improves and guardrails hold.
Rollback by switching current pointers and invalidating caches.
