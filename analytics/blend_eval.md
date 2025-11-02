# Blend Weight Offline Evaluation (RT-3A)

This harness replays recent user interactions and re-runs the recommendation engine against multiple blend configurations so we can compare hit-rate, MRR, and variety **before** shipping new weights.

## Command

```bash
# From repo root (requires DATABASE_URL and API envs, e.g. api/.env.test)
go run ./api/cmd/blend_eval \
  -namespace default \
  -k 20 \
  -limit 500 \
  -lookback 720h \
  -configs analytics/blend_eval_configs.example.yaml
```

Flags:

| Flag | Description |
|------|-------------|
| `-namespace` | Namespace to sample (defaults to `default`). |
| `-k` | Recommendation list length for evaluation (`20`). |
| `-limit` | Maximum user samples to evaluate (`200`). |
| `-min-events` | Minimum interactions per user (`5`). |
| `-lookback` | Event lookback window (Go duration). |
| `-configs` | Optional YAML file describing candidate blends. |

If `-configs` is omitted the tool runs three baseline presets (`baseline`, `pop-heavy`, `embed-heavy`). Results are printed as a table with hit-rate, mean reciprocal rank, average rank, coverage, and failure counts.

## Config File Structure

```yaml
configs:
  - name: baseline
    description: Current production mix
    alpha: 0.30
    beta: 0.50
    gamma: 0.20
  - name: explore_content
    alpha: 0.20
    beta: 0.30
    gamma: 0.50
```

## How It Works

1. Samples users whose latest event occurs within the lookback window and have at least `min-events` interactions.
2. For each candidate blend, replays the engine (`algorithm.Engine`) with the specified weights via request overrides.
3. Records whether the final interaction (`holdout` item) appears in the top‑K list and aggregates:
   - Hit rate @K
   - Mean reciprocal rank
   - Average rank when hit
   - Catalog coverage (`unique_items / (samples * K)`)
   - Response/engine failures

## Next Steps

- Persist results to cloud storage for longitudinal tracking.
- Wire into CI with a headless Postgres fixture for regression gating.
- Extend scoring to also capture diversity metrics (brand/category entropy) alongside blend performance.
