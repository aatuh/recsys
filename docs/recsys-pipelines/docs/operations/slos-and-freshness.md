---
diataxis: explanation
tags:
  - recsys-pipelines
---
# SLOs and freshness
This page explains SLOs and freshness and how it fits into the RecSys suite.


## Freshness definition

A surface is "fresh" if the manifest for (tenant, surface) was updated within
an expected time window.

Example daily schedule:

- Run for previous UTC day at 01:00 UTC
- Expect publish to finish by 01:30 UTC

## What to measure

At minimum:

- last successful publish timestamp per tenant/surface
- validation failures count
- limit exceeded failures count
- runtime per job

## Alert suggestions

- Stale manifest: no update within 2x schedule interval
- Persistent validation failures
- Persistent limit exceeded

## Where to find the signal in local mode

- Manifest `updated_at`:

  `.out/registry/current/<tenant>/<surface>/manifest.json`

## Read next

- Operate pipelines daily: [How-to: Operate pipelines daily](../how-to/operate-daily.md)
- Schedule pipelines: [How-to: Schedule pipelines with CronJob](../how-to/schedule-pipelines.md)
- Stale artifacts runbook: [Runbook: Stale artifacts](runbooks/stale-artifacts.md)
- Pipeline failed runbook: [Runbook: Pipeline failed](runbooks/pipeline-failed.md)
