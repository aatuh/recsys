# CI gates: using recsys-eval in automation

## Who this is for
Engineers wiring recsys-eval into CI/CD or scheduled pipelines.

## What you will get
- A practical gating pattern
- How to use exit codes
- How to store artifacts and compare runs

## The pattern: validate -> run -> store report -> gate

1) Validate data (optional but recommended)
2) Run evaluation
3) Upload report artifact
4) Fail the pipeline if gates fail

Example (tiny dataset gate used in CI):

```bash
recsys-eval run \
  --mode offline \
  --dataset configs/examples/dataset.jsonl.yaml \
  --config configs/eval/offline.ci.yaml \
  --output /tmp/offline_report.json \
  --baseline testdata/golden/offline.json
```

## Exit codes

recsys-eval is designed to be automation-friendly:
- configuration or schema errors should fail fast
- gate failures should fail deterministically

Recommended practice:
- treat "invalid input" differently from "metric regression"

If your build supports a decision artifact:
- fail if decision != ship
- attach decision.json and report.json to the build

## Artifact storage

Store:
- report.json
- effective config (or config hash)
- dataset fingerprint / window
- the exact binary version (build info)

This is what makes runs auditable.

## Golden tests vs production gates

Golden tests:
- use tiny datasets
- protect behavior and output stability

Production gates:
- use real logs
- protect business impact and safety

Do not confuse the two. Use both.
