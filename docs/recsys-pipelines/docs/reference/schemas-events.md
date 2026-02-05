
# Event schemas

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

- Start here: [`start-here.md`](../start-here.md)
- Add event field: [`how-to/add-event-field.md`](../how-to/add-event-field.md)
- Data lifecycle: [`explanation/data-lifecycle.md`](../explanation/data-lifecycle.md)
- Run a backfill: [`how-to/run-backfill.md`](../how-to/run-backfill.md)
