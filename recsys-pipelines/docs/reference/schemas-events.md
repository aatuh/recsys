
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
