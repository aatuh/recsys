---
diataxis: how-to
tags:
  - recsys-pipelines
---
# How-to: Run incremental pipelines

Incremental runs use a checkpoint so you can process only new days.

## Prerequisites

- `checkpoint_dir` configured (defaults to `.out/checkpoints`)
- `raw_source` configured

## Run

```bash
recsys-pipelines run \
  --config configs/env/local.json \
  --tenant demo \
  --surface home \
  --end 2026-02-01 \
  --incremental
```

First run: pass `--start` once to seed the checkpoint:

```bash
recsys-pipelines run \
  --config configs/env/local.json \
  --tenant demo \
  --surface home \
  --start 2026-01-01 \
  --end 2026-01-07 \
  --incremental
```

After each successful day, the checkpoint is updated automatically.

## Read next

- Operate pipelines daily: [How-to: Operate pipelines daily](operate-daily.md)
- Schedule pipelines: [How-to: Schedule pipelines with CronJob](schedule-pipelines.md)
- SLOs and freshness: [SLOs and freshness](../operations/slos-and-freshness.md)
- Output layout: [Output layout (local filesystem)](../reference/output-layout.md)
- Config reference: [Config reference](../reference/config.md)
