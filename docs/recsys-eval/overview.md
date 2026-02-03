# recsys-eval

recsys-eval turns recommendation logs into reports that tell you whether a
recommender change is better, worse, or unclear - globally and per segment -
with guardrails.

If you only read one thing: read [`Concepts`](docs/concepts.md).

## Who this is for

- Engineers shipping recommender changes
- Analysts and DS folks validating impact
- Platform teams wiring evaluation into CI
- Anyone who wants a clear "ship / hold / rollback" decision trail

## What you get

- Offline evaluation (fast regression gate)
- Experiment analysis (A/B from production logs)
- Off-policy evaluation (OPE) when experiments are hard
- Interleaving analysis for sensitive ranker comparisons
- JSON/Markdown/HTML reports + optional decision artifact

## Quick start (JSONL)

1) Validate your inputs (recommended):

```bash
recsys-eval validate \
  --schema exposure.v1 \
  --input testdata/datasets/tiny/exposures.jsonl
```

1) Run an evaluation (choose one mode):

```bash
# Offline evaluation
recsys-eval run \
  --mode offline \
  --dataset configs/examples/dataset.jsonl.yaml \
  --config configs/eval/offline.default.yaml \
  --output /tmp/offline_report.json

# Markdown report
recsys-eval run \
  --mode offline \
  --dataset configs/examples/dataset.jsonl.yaml \
  --config configs/eval/offline.default.yaml \
  --output /tmp/offline_report.md \
  --output-format markdown

# Experiment analysis
recsys-eval run \
  --mode experiment \
  --dataset configs/examples/dataset.jsonl.yaml \
  --config configs/eval/experiment.default.yaml \
  --output /tmp/experiment_report.json

# Offline evaluation (signals sample dataset)
recsys-eval run \
  --mode offline \
  --dataset configs/examples/dataset.signals.yaml \
  --config configs/eval/offline.signals.yaml \
  --output /tmp/offline_signals_report.json

# Off-policy evaluation (OPE)
recsys-eval run \
  --mode ope \
  --dataset configs/examples/dataset.jsonl.yaml \
  --config configs/eval/ope.default.yaml \
  --output /tmp/ope_report.json

# Interleaving analysis
recsys-eval run \
  --mode interleaving \
  --dataset configs/examples/dataset.interleaving.jsonl.yaml \
  --config configs/eval/interleaving.default.yaml \
  --output /tmp/interleaving_report.json
```

## Outputs

The primary output is a JSON report that conforms to api/schemas/report.v1.json.
You can also emit Markdown or HTML summaries for sharing.
It always includes:

- run_id
- mode
- created_at
- version
- summary

Mode-specific sections are included when relevant:
offline, experiment, ope, interleaving, aa_check.

Optionally, some modes can emit a decision artifact that conforms to
api/schemas/decision.v1.json.

## Read next

- [`Concepts`](docs/concepts.md): what the system does and how to think about it
- [`Data contracts`](docs/data_contracts.md): what your inputs must look like
- [`Interpreting results`](docs/interpreting_results.md): how to read reports and make decisions

Company-grade additions:

- [`Integration`](docs/integration.md): how to emit logs from a serving system
- [`CI gates`](docs/ci_gates.md): exit codes, gating, and recommended pipelines
- [`Scaling`](docs/scaling.md): large datasets and stream mode
- [`Runbooks`](docs/runbooks.md) and [`Troubleshooting`](docs/troubleshooting.md): debug and operate it
- [`OPE`](docs/ope.md) and [`Interleaving`](docs/interleaving.md): deeper dives
- [`Architecture`](docs/architecture.md): extension points and how to add features

---

## Releases

Tag releases with the module prefix, e.g. `recsys-eval/v0.2.0`.
