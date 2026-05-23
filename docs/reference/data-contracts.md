# Data Contracts

## Who this is for

Data engineers, product analysts, backend developers, and evaluators joining recommendation exposure data to outcomes.

## What you will get

- The canonical schema file locations.
- Join keys for exposure/outcome/assignment data.
- Privacy and retention expectations for EU-baseline deployments.
- A minimal event shape example.

## Canonical schema sources

| Contract | Source |
| --- | --- |
| Evaluation exposure | `recsys-eval/schemas/exposure.v1.json` |
| Evaluation outcome | `recsys-eval/schemas/outcome.v1.json` |
| Evaluation assignment | `recsys-eval/schemas/assignment.v1.json` |
| Evaluation ranklist | `recsys-eval/schemas/ranklist.v1.json` |
| Evaluation report | `recsys-eval/schemas/report.v1.json` |
| Evaluation decision | `recsys-eval/schemas/decision.v1.json` |
| Pipeline exposure event | `recsys-pipelines/schemas/events/exposure.v1.json` |
| Artifact manifest | `recsys-pipelines/schemas/artifacts/manifest.v1.json` |

## Join model

| Data | Required keys |
| --- | --- |
| Exposure | `request_id`, `user_id` or pseudonymous equivalent, timestamp, displayed items. |
| Outcome | `request_id`, user identifier, `item_id`, event type, timestamp. |
| Assignment | `experiment_id`, `variant`, `request_id`, user identifier, timestamp. |
| Manifest | tenant, surface, current artifact pointers, update timestamp. |

Stable `request_id` values are the most important reconstruction key. Preserve them in service logs, client telemetry,
exposure events, and evaluation datasets.

## Minimal exposure example

```json
{
  "request_id": "req_123",
  "user_id": "anon_456",
  "ts": "2026-05-23T12:00:00Z",
  "items": [
    {"item_id": "item_1", "rank": 1},
    {"item_id": "item_2", "rank": 2}
  ],
  "context": {
    "surface": "home",
    "device": "web"
  }
}
```

## Privacy posture

- Use pseudonymous IDs in request, exposure, assignment, and outcome data.
- Do not place direct PII in item IDs, user IDs, context fields, logs, reports, or exported datasets.
- Document retention for exposure and outcome data before production.
- Treat evaluation datasets as sensitive because they reconstruct user behavior even when identifiers are pseudonymous.
