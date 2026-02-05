
# Tutorial: Run in job-per-step mode

Some teams prefer orchestration where each step runs as a separate job
(Airflow, K8s CronJobs, etc.). This repo includes job binaries:

- `job_ingest`
- `job_validate`
- `job_popularity`
- `job_cooc`
- `job_publish`

## Example: one day

```bash
make build

./bin/job_ingest --config configs/env/local.json --tenant demo --surface home \
  --start 2026-01-01 --end 2026-01-01

./bin/job_validate --config configs/env/local.json --tenant demo --surface home \
  --start 2026-01-01 --end 2026-01-01

./bin/job_popularity --config configs/env/local.json --tenant demo --surface home \
  --segment '' --start 2026-01-01 --end 2026-01-01

./bin/job_cooc --config configs/env/local.json --tenant demo --surface home \
  --segment '' --start 2026-01-01 --end 2026-01-01

./bin/job_publish --config configs/env/local.json --tenant demo --surface home \
  --segment '' --start 2026-01-01 --end 2026-01-01
```

## Why split jobs?

- Different compute profiles per step
- Independent retries
- Separate scaling policies

## Read next

- SLOs and freshness: [`operations/slos-and-freshness.md`](../operations/slos-and-freshness.md)
- Schedule pipelines: [`how-to/schedule-pipelines.md`](../how-to/schedule-pipelines.md)
- Debug failures: [`how-to/debug-failures.md`](../how-to/debug-failures.md)
- Start here: [`start-here.md`](../start-here.md)
