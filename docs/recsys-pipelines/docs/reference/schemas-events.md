---
diataxis: reference
tags:
  - recsys-pipelines
---
# Event schemas
This page is the canonical reference for Event schemas.


## ExposureEventV1

Schema file:

- `schemas/events/exposure.v1.json`

Format:

- JSON Lines (one JSON object per line)

Required fields:

- `v`: must be 1
- `ts`: RFC3339 timestamp
- `tenant`: string
- `surface`: string
- `session_id`: string
- `item_id`: string

Optional fields:

- `user_id`
- `request_id`
- `rank`

See also: `testdata/events/`.

## Read next

- Start here: [Start here](../start-here.md)
- Add event field: [How-to: Add a new field to exposure events](../how-to/add-event-field.md)
- Data lifecycle: [Data lifecycle](../explanation/data-lifecycle.md)
- Run a backfill: [How-to: Run a backfill safely](../how-to/run-backfill.md)
