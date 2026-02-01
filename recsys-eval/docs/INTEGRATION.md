# Integration: how to produce the inputs

## Who this is for
Backend / platform engineers wiring recsys-eval into a real recommender stack.

## What you will get
- What you need to log in your serving system
- How to keep IDs stable and privacy-safe
- A minimal logging plan for each mode

## The one rule: always log exposures

If you want to measure recommendations, you must log "what you showed".
Clicks without exposures are not evaluatable.

## Exposure logging (recommended fields)

At recommendation time (serving):
- request_id: unique per recommendation request
- tenant, surface: required for segmentation
- user_id or session_id: pseudonymized
- timestamp (ISO-8601)
- the ranked list of items (item_id + rank)
- optional: latency_ms, model_version, config_version, algo_version
- optional for deeper analysis: per-item scores and reasons (if you have them)

Minimal JSONL exposure record:
```json
{
  "request_id": "req_123",
  "tenant": "demo",
  "surface": "home",
  "user_id": "u_hash_...",
  "timestamp": "2026-01-27T12:00:00Z",
  "items": [{"item_id": "A", "rank": 1}]
}
```

## Outcome logging

After exposure, when the user acts:
- request_id (same one)
- event type: click, purchase, etc.
- item_id (the item clicked/converted)
- timestamp

If you have revenue or value, log it. If you do not, do not invent it.

## Assignment logging (experiments)

When you run an experiment:
- assignment should be deterministic and consistent
- log control vs candidate in a way you can audit

Minimum:
- request_id
- experiment_id
- variant

## OPE logging (advanced)

If you want OPE:
- you must log propensities
- you must define what policy produced the logged exposures

This is easy to get wrong. Read docs/OPE.md before attempting.

## Privacy and IDs

Guidelines:
- never log raw PII (email, phone)
- hash or pseudonymize user identifiers
- be consistent: the same user should map to the same pseudonymous ID

## "Minimal viable integration" by mode

Offline:
- exposures + outcomes
- no assignments needed

Experiment:
- exposures + outcomes + assignments

Interleaving:
- ranker_a results + ranker_b results + outcomes

OPE:
- exposures + outcomes + propensities (hard requirement)

## Operational tip
Start with the tiny dataset shipped in testdata. If you cannot make your
production logs look like that, you will struggle later.
