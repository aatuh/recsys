---
diataxis: tutorial
tags:
  - recsys-pipelines
---
# Tutorial: Run locally (filesystem mode)

This tutorial assumes you want to run the pipeline on the included tiny
dataset.

## 1) Build

```bash
make test
make build
```

## 2) Run one day

```bash
./bin/recsys-pipelines run \
  --config configs/env/local.json \
  --tenant demo \
  --surface home \
  --start 2026-01-01 \
  --end 2026-01-01
```

## 3) Inspect outputs

```bash
find .out -type f | sort
cat .out/registry/current/demo/home/manifest.json
```

You should see:

- canonical events under `.out/canonical/demo/home/exposures/`
- versioned blobs under `.out/objectstore/demo/home/...`
- a manifest pointer under `.out/registry/current/demo/home/manifest.json`

## 4) Prove idempotency

Run the same command again. Output should not change.

A smoke script exists:

```bash
make smoke
```

## What you learned

- How to build and run the pipeline locally
- Where the outputs land
- Why reruns are safe

## Read next

- Job-per-step mode: [Run in job-per-step mode](job-mode.md)
- Artifacts and versioning: [Artifacts and versioning](../explanation/artifacts-and-versioning.md)
- Output layout: [Output layout (local filesystem)](../reference/output-layout.md)
- Start here: [Start here](../start-here.md)
