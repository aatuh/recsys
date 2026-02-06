---
tags:
  - reference
  - cli
  - recsys-eval
  - developer
  - ops
---

# CLI: recsys-eval

## Who this is for

- Engineers running `recsys-eval` locally or in CI
- Teams implementing evaluation gates (ship/hold/fail) based on reports

## What you will get

- The canonical `recsys-eval` commands, flags, and exit codes
- Copy/paste examples for local runs and CI gating

## Build/install

From repo root:

```bash
(cd recsys-eval && make build)
```

Binary:

- `recsys-eval/bin/recsys-eval`

## Commands

### `recsys-eval run`

Runs one evaluation mode and writes a report.

Required flags:

- `--dataset <path.yaml>`: dataset config (YAML)
- `--config <path.yaml>`: eval config (YAML)
- `--output <path>`: output report path

Common optional flags:

- `--mode <offline|experiment|ope|interleaving|aa-check>`: overrides `mode:` in config
- `--output-format <json|markdown|html>`: default is `json`
- `--baseline <path.json>`: offline mode baseline report (JSON) for comparisons
- `--experiment-id <id>`: experiment mode override

### `recsys-eval validate`

Validates a JSONL file against a schema.

- `--schema <name-or-path>`: schema name (like `exposure.v1`) or a path to a `.json` schema file
- `--input <path.jsonl>`: JSONL file to validate

Notes:

- If `--schema` does not end with `.json`, `recsys-eval` resolves it as `schemas/<schema>.json`.
  That means `--schema exposure.v1` works when run from the `recsys-eval/` directory.

### `recsys-eval version`

Prints the CLI version.

## Exit codes

- `0`: success (and `ship` decision in experiment mode)
- `1`: command failed (invalid input/config, validation errors, runtime errors)
- `2`: experiment decision is `hold`
- `3`: experiment decision is `fail`

## Examples

### Local: validate + offline report

```bash
(cd recsys-eval && ./bin/recsys-eval validate --schema exposure.v1 --input /tmp/exposures.jsonl)
(cd recsys-eval && ./bin/recsys-eval validate --schema outcome.v1 --input /tmp/outcomes.jsonl)

recsys-eval/bin/recsys-eval run \
  --mode offline \
  --dataset /tmp/dataset.yaml \
  --config /tmp/eval.yaml \
  --output /tmp/recsys_eval_report.md \
  --output-format markdown
```

### CI: experiment gate (ship/hold/fail)

```bash
set +e
recsys-eval/bin/recsys-eval run \
  --mode experiment \
  --dataset /tmp/dataset.yaml \
  --config /tmp/experiment.yaml \
  --output /tmp/recsys_eval_report.json \
  --output-format json
code="$?"
set -e

case "$code" in
  0) echo "decision=ship" ;;
  2) echo "decision=hold" ; exit 0 ;;
  3) echo "decision=fail" ; exit 1 ;;
  *) echo "decision=error" ; exit 1 ;;
esac
```

## Read next

- Default evaluation pack: [`recsys-eval/docs/default-evaluation-pack.md`](../../recsys-eval/docs/default-evaluation-pack.md)
- Run eval and ship decisions: [`how-to/run-eval-and-ship.md`](../../how-to/run-eval-and-ship.md)
- Data contracts (schemas + examples): [`reference/data-contracts/index.md`](../data-contracts/index.md)
