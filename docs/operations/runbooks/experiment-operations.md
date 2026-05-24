# Runbook: Experiment Operations

Use this when launching, holding, or rolling back a recommendation experiment.

## Configure

Set `EXPERIMENT_ASSIGNMENT_ENABLED=true`, provide `EXPERIMENT_ASSIGNMENT_SALT`, and optionally define lifecycle rules:

```bash
EXPERIMENT_CONFIG_JSON='[
  {
    "id": "home-ranker-v2",
    "enabled": true,
    "surface": "home",
    "traffic_percent": 25,
    "variants": ["A", "B"],
    "starts_at": "2026-01-01T00:00:00Z",
    "ends_at": "2026-02-01T00:00:00Z"
  }
]'
```

Assignments are deterministic for the same experiment ID and subject. If an experiment is disabled, outside its time
window, outside its traffic allocation, or sent for the wrong surface, the API leaves the variant empty.

## Launch Checks

1. Confirm clients send the intended `experiment.id`.
2. Confirm exposure logs include `experiment_id`, `experiment_variant`, `request_id`, tenant, surface, and subject hash.
3. Verify `/metrics` shows stable errors, latency, warning counts, and empty recommendation rate.
4. Run a small traffic allocation before increasing blast radius.

## Hold Or Roll Back

Hold when assignment joins, sample ratio, or exposure/outcome schemas are not trustworthy. Roll back when guardrails
regress materially or KPI movement is clearly negative.

Rollback options:

- Set the experiment definition `enabled` field to `false` and redeploy config.
- Reduce `traffic_percent` to `0`.
- Revert the ranking config, rules, artifact manifest, or service image involved in the experiment.

## Evidence To Keep

- Experiment config JSON and deployment time.
- Assignment/export paths and row counts.
- `recsys-eval` report path and hash.
- Guardrail summary: errors, latency, empty recommendations, and warnings.
- Rollback lever used, if any.
