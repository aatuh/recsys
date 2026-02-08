---
diataxis: how-to
tags:
  - recsys-pipelines
---
# How-to: Schedule pipelines with CronJob

This project ships CLI jobs; schedule them with Kubernetes CronJobs or system cron.

## Example Kubernetes CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: recsys-pipelines-nightly
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: recsys-pipelines
              image: ghcr.io/aatuh/recsys-pipelines:latest
              args:
                - "run"
                - "--config"
                - "/etc/recsys/config.json"
                - "--tenant"
                - "demo"
                - "--surface"
                - "home"
                - "--end"
                - "2026-02-01"
                - "--incremental"
              volumeMounts:
                - name: recsys-config
                  mountPath: /etc/recsys
          restartPolicy: OnFailure
          volumes:
            - name: recsys-config
              configMap:
                name: recsys-pipelines-config
```

Use `--incremental` for daily runs and `--start/--end` for backfills.

## Read next

- Operate pipelines daily: [How-to: Operate pipelines daily](operate-daily.md)
- Run incremental: [How-to: Run incremental pipelines](run-incremental.md)
- SLOs and freshness: [SLOs and freshness](../operations/slos-and-freshness.md)
- Pipeline failed runbook: [Runbook: Pipeline failed](../operations/runbooks/pipeline-failed.md)
- Config reference: [Config reference](../reference/config.md)
