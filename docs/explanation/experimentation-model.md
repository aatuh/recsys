# Experimentation model (A/B, interleaving, OPE)

## Who this is for

- Product and stakeholders who need a clear “how do we measure lift?” story
- Engineers wiring evaluation into CI and production workflows
- Recommendation engineers choosing between A/B, interleaving, and OPE

## What you will get

- A decision guide for which evaluation mode to use (and when)
- The instrumentation required for each mode (what to log)
- How `recsys-service` supports experiment metadata and deterministic bucketing
- Common failure modes (SRM, broken joins, confounded tests)

## The key idea: measure with logs

Every evaluation mode in this suite is built on the same foundation:

- **Expose**: what you showed (ranked list)
- **Outcome**: what the user did (click/conversion)
- **Correlate**: join by `request_id`

If exposures or `request_id` are missing, everything else becomes guesswork.

See: [`explanation/exposure-logging-and-attribution.md`](exposure-logging-and-attribution.md)

## Choosing a mode (what to use when)

Use this as your default decision guide:

| Goal | Mode | What you need | What you get |
| --- | --- | --- | --- |
| Regression gate | Offline evaluation | exposures + outcomes | ranking metrics (NDCG/Recall/etc.) |
| KPI lift (shipping) | Experiment (A/B) | exposures + outcomes + assignments | KPI deltas + guardrails + SRM |
| Ranker comparison | Interleaving | ranklist A + ranklist B + outcomes | win rate + significance |
| Estimate (no randomize) | OPE | exposures + outcomes + propensities | IPS/SNIPS/DR + diagnostics |

Notes:

- Offline metrics are excellent for “did we break something?”, but they are not a replacement for measuring KPI lift.
- OPE is powerful but easy to get wrong; treat it as advanced and validate assumptions carefully.

## Required instrumentation (minimal)

The suite uses `recsys-eval` data contracts. At minimum:

- **Exposure** (`exposure.v1` / eval JSONL): `request_id`, `user_id`, `ts`, `items[]`
- **Outcome** (`outcome.v1`): `request_id`, `user_id`, `item_id`, `event_type`, `ts`

Mode-specific:

- **Experiment (A/B)**: assignment stream (`assignment.v1`) with `experiment_id`, `variant`, `request_id`, `user_id`, `ts`
- **Interleaving**: rank lists (`ranklist.v1`) for ranker A and ranker B (same `request_id` join key)
- **OPE**: propensities on each exposed item (`propensity` fields on exposure items)

Full schemas: [`reference/data-contracts/eval-events.md`](../reference/data-contracts/eval-events.md)

## Experiment metadata in `recsys-service`

The recommend API accepts optional experiment metadata:

```json
{
  "surface": "home",
  "k": 10,
  "user": { "user_id": "u_123" },
  "experiment": { "id": "exp_home_rank_v2", "variant": "B" }
}
```

What the service does with it:

- The experiment is included in exposure logging (when `EXPOSURE_LOG_FORMAT=eval_v1`) as `experiment_id` and
  `experiment_variant` context keys.
- The service does **not** change ranking behavior based on `experiment.variant`. Your application (or an experiment
  platform) must decide what differs between control and candidate (for example: `algorithm`, `weights`, or an upstream
  candidate set).

### Deterministic variant assignment (optional)

If you provide an experiment ID but omit the variant:

- and `EXPERIMENT_ASSIGNMENT_ENABLED=true`
- and you provide at least one stable identifier (`user_id`, `session_id`, or `anonymous_id`)

then the service assigns a deterministic variant during request normalization (see `POST /v1/recommend/validate`).

Configure:

- `EXPERIMENT_DEFAULT_VARIANTS` (default: `A,B`)
- `EXPERIMENT_ASSIGNMENT_SALT` (recommended: set this; defaults to `EXPOSURE_HASH_SALT`)

This feature is primarily for **consistent logging** and debugging; it is not a full experimentation platform.

## Getting `assignment.v1` events (practical options)

`recsys-eval` experiment analysis expects a separate assignment stream.

You have two good options:

1. **If you already have an experimentation platform**: export its assignment logs into `assignment.v1`.
2. **If you use `recsys-service` exposure logs**: derive assignments from exposure records (when experiment context is
   present):

   ```bash
   jq -c '
     select(.context.experiment_id and .context.experiment_variant) |
     {
       experiment_id: .context.experiment_id,
       variant: .context.experiment_variant,
       request_id: .request_id,
       user_id: .user_id,
       ts: .ts,
       context: {
         tenant_id: .context.tenant_id,
         surface: .context.surface,
         segment: .context.segment
       }
     }
   ' exposures.eval.jsonl > assignments.jsonl
   ```

## Common experiment failure modes

- **Broken joins** (missing/mismatched `request_id`)
  - Symptom: low join rate, unstable metrics.
  - Fix: follow the join rules in [`reference/data-contracts/join-logic.md`](../reference/data-contracts/join-logic.md).
- **SRM (sample ratio mismatch)**
  - Symptom: recsys-eval report warns that buckets are imbalanced.
  - Fix: ensure deterministic assignment and stable subject IDs; avoid platform-specific bucketing bugs.
- **Confounded experiments**
  - Symptom: variant B is “better”, but you changed multiple things at once.
  - Fix: keep the treatment minimal (one meaningful change), and record config/rules/algo versions in logs.

## Read next

- Run eval and ship (suite workflow): [`how-to/run-eval-and-ship.md`](../how-to/run-eval-and-ship.md)
- recsys-eval concepts (modes and pitfalls): [`recsys-eval/docs/concepts.md`](../recsys-eval/docs/concepts.md)
- recsys-eval interleaving and OPE: [`recsys-eval/docs/interleaving.md`](../recsys-eval/docs/interleaving.md),
  [`recsys-eval/docs/ope.md`](../recsys-eval/docs/ope.md)
