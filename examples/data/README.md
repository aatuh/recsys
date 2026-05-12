# Demo data (synthetic)

This folder contains permissively licensed synthetic datasets that can run the
pipelines and proof-kit flows end-to-end without external inputs.

## Available datasets

- `tiny/`: minimal pipeline fixture used by older demos and smoke tests.
- `ecommerce-mini/`: synthetic ecommerce fixture used by the commercial proof
  kit. It includes pipeline raw exposures, eval fixtures, and a small product
  catalog.

## Schema: exposure.jsonl

Each line is a JSON object:

- `v` (int): schema version
- `ts` (RFC3339): event timestamp
- `tenant` (string): tenant external id (e.g., `demo`)
- `surface` (string): surface/namespace (e.g., `home`)
- `session_id` (string): session identifier
- `item_id` (string): item identifier
- `rank` (int): position shown to the user

Example:

```json
{"v":1,"ts":"2026-01-01T08:00:00Z","tenant":"demo","surface":"home","session_id":"s1","item_id":"A","rank":1}
```

## Generation method

This dataset is hand-authored synthetic data intended to be deterministic and
small enough for local demos.
