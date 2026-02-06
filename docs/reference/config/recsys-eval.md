---
tags:
  - reference
  - config
  - recsys-eval
  - developer
  - ml
---

# recsys-eval configuration

## Who this is for

- RecSys engineers running offline regression gates in CI
- Developers validating instrumentation by producing a first report from logs

## What you will get

- The evaluation config schema (what `--config` expects)
- The dataset wiring schema (what `--dataset` expects)
- Copy/paste examples you can adapt

## Reference

`recsys-eval run` takes two YAML files:

- `--dataset`: where exposures/outcomes/assignments come from (sources and joins happen in the tool)
- `--config`: what evaluation to run (mode, metrics, gates, guardrails)

Important:

- YAML parsing is **strict**: unknown fields fail fast.
- Output format is a CLI flag (`--output-format`), not a config field.

### Dataset config (`--dataset`)

Top-level keys:

| Key | Required | Meaning |
| --- | --- | --- |
| `exposures` | offline/experiment/ope | Exposure source (what was shown). |
| `outcomes` | offline/experiment/ope | Outcome source (what the user did). |
| `assignments` | experiment/aa-check | Experiment assignment source (variant per request/user). |
| `interleaving` | interleaving | Special wiring for interleaving (ranker A/B lists + outcomes). |

Source config:

| Key | Required | Meaning |
| --- | --- | --- |
| `type` | yes | `jsonl`, `postgres`, or `duckdb`. |
| `path` | jsonl | Path to a JSONL file. |
| `dsn` | postgres/duckdb | DB DSN. |
| `query` | postgres/duckdb | Query that returns JSON rows. |

### Evaluation config (`--config`)

Top-level keys:

| Key | Default | Meaning |
| --- | --- | --- |
| `mode` | required | `offline`, `experiment`, `ope`, `interleaving`, `aa-check`. |
| `offline` | empty | Offline regression metrics and gates (used in `offline` mode). |
| `experiment` | empty | Experiment analysis and guardrails (used in `experiment` and `aa-check` modes). |
| `ope` | defaults set | Off-policy evaluation settings (used in `ope` mode). |
| `interleaving` | defaults set | Interleaving algorithm settings (used in `interleaving` mode). |
| `scale` | `memory` | `memory`, `stream`, or `duckdb` mode for large datasets. |
| `artifacts` | empty | Optional artifact/manifest resolution metadata (for report context). |

Defaults applied by the tool:

- `scale.mode`: defaults to `memory`
- `ope.reward_event`: defaults to `click`
- `ope.unit`: defaults to `request`
- `ope.reward_aggregation`: defaults to `sum`
- `ope.min_propensity`: defaults to `1e-6`
- `interleaving.algorithm`: defaults to `team_draft`
- `interleaving.seed`: defaults to `42`

Offline mode requirements:

- `offline.metrics` is required (at least one metric spec)

## Examples

### Minimal dataset config (JSONL)

```yaml
exposures:
  type: jsonl
  path: /tmp/exposures.eval.jsonl
outcomes:
  type: jsonl
  path: /tmp/outcomes.eval.jsonl
assignments:
  type: jsonl
  path: /tmp/assignments.eval.jsonl
```

### Minimal offline config (gate in CI)

```yaml
mode: offline
offline:
  metrics:
    - name: precision
      k: 10
    - name: recall
      k: 10
  slice_keys: ["tenant", "surface"]
  gates:
    - metric: precision@10
      max_drop: 0.001
```

### Minimal experiment config (guardrails + primary metrics)

```yaml
mode: experiment
experiment:
  experiment_id: "exp_123"
  control_variant: "A"
  primary_metrics: ["ctr", "conversion_rate"]
  slice_keys: ["tenant", "surface"]
  guardrails:
    max_latency_p95_ms: 300
    max_error_rate: 0.01
    max_empty_rate: 0.02
```

## Read next

- CLI usage and exit codes: [`reference/cli/recsys-eval.md`](../cli/recsys-eval.md)
- How-to run eval and ship decisions: [`how-to/run-eval-and-ship.md`](../../how-to/run-eval-and-ship.md)
- Default evaluation pack: [`recsys-eval/docs/default-evaluation-pack.md`](../../recsys-eval/docs/default-evaluation-pack.md)
