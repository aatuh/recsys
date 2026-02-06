---
tags:
  - reference
  - data-contracts
  - recsys-eval
  - developer
  - evaluation
---

# recsys-eval event schemas (v1)

## Who this is for

- Developers producing exposure/outcome/assignment JSONL files
- Data engineers validating that logs meet strict schema requirements

## What you will get

- Required fields and allowed optional fields for each event type
- JSONL examples you can copy/paste
- Guidance to avoid validation failures

`recsys-eval validate` uses strict JSON schemas with
`additionalProperties: false`. If your JSONL includes extra keys or misses
required fields, validation fails even if `recsys-eval run` produced a report.

Schema sources (repo):

- `recsys-eval/schemas/exposure.v1.json`
- `recsys-eval/schemas/outcome.v1.json`
- `recsys-eval/schemas/assignment.v1.json`

## exposure.v1 (required fields)

Required:

- `request_id` (string, non-empty)
- `user_id` (string, non-empty)
- `ts` (RFC3339 date-time)
- `items` (array of `{ item_id, rank }`)

Allowed optional fields:

- `context` (object of string values)
- `latency_ms` (number >= 0)
- `error` (boolean)
- `items[].propensity`, `items[].logging_propensity`, `items[].target_propensity`

Examples (JSONL; one object per line):

```json
{"request_id":"req-1","user_id":"user-1","ts":"2026-01-30T12:00:00Z","items":[{"item_id":"item-1","rank":1},{"item_id":"item-2","rank":2}],"context":{"surface":"home","tenant_id":"demo"}}
{"request_id":"req-2","user_id":"user-2","ts":"2026-01-30T12:05:00Z","items":[{"item_id":"sku-9","rank":1,"propensity":0.42,"logging_propensity":0.5,"target_propensity":0.6},{"item_id":"sku-3","rank":2}],"context":{"surface":"home","segment":"default","tenant_id":"demo","locale":"en-US"},"latency_ms":12.4,"error":false}
```

## outcome.v1 (required fields)

Required:

- `request_id` (string, non-empty)
- `user_id` (string, non-empty)
- `item_id` (string, non-empty)
- `event_type` (`click` or `conversion`)
- `ts` (RFC3339 date-time)

Allowed optional fields:

- `value` (number)

Examples (JSONL; one object per line):

```json
{"request_id":"req-1","user_id":"user-1","item_id":"item-2","event_type":"click","ts":"2026-01-30T12:00:03Z"}
{"request_id":"req-1","user_id":"user-1","item_id":"item-2","event_type":"conversion","value":49.90,"ts":"2026-01-30T12:00:20Z"}
{"request_id":"req-2","user_id":"user-2","item_id":"sku-9","event_type":"click","ts":"2026-01-30T12:05:02Z"}
```

## assignment.v1 (required fields)

Required:

- `experiment_id` (string, non-empty)
- `variant` (string, non-empty)
- `request_id` (string, non-empty)
- `user_id` (string, non-empty)
- `ts` (RFC3339 date-time)

Allowed optional fields:

- `context` (object of string values)

Examples (JSONL; one object per line):

```json
{"experiment_id":"exp-1","variant":"A","request_id":"req-1","user_id":"user-1","ts":"2026-01-30T12:00:00Z","context":{"surface":"home"}}
{"experiment_id":"exp-1","variant":"B","request_id":"req-2","user_id":"user-2","ts":"2026-01-30T12:05:00Z","context":{"surface":"home","tenant_id":"demo"}}
```

## How to avoid validation failures

- Ensure every record contains **exactly** the required fields (no extras).
- Use RFC3339 timestamps for `ts`.
- `user_id` must be non-empty (use a stable pseudonymous ID if needed).
- Join key is **request_id**: outcomes and assignments must use the same

  `request_id` as the exposure.

## Service exposure logs vs eval schema

`recsys-service` logs exposures in its native format by default. To emit
`exposure.v1` directly for recsys-eval, set:

```bash
EXPOSURE_LOG_FORMAT=eval_v1
```

If you keep `service_v1`, you must transform logs before running
`recsys-eval validate`.
