---
tags:
  - ops
  - overview
---

# Operations

This section is for **running RecSys in production**: performance, readiness, and on-call runbooks.

## Who this is for

- SREs and on-call engineers running RecSys in production
- Engineering teams sizing capacity and validating production readiness

## What you will get

- A production readiness checklist and baseline benchmarks
- Failure-mode diagnosis and safe remediations (with runbook links)
- The first pages to open when the service is not ready or recommendations go empty

## Quick paths

<div class="grid cards" markdown>

- **[Performance & capacity](performance-and-capacity.md)**  
  Sizing guidance and performance expectations.
- **[Baseline benchmarks](baseline-benchmarks.md)**  
  Reproducible “anchor numbers” and a template to record your own runs.
- **[Production readiness checklist](production-readiness-checklist.md)**  
  Pre-flight checks before you go live.
- **[Failure modes & diagnostics](failure-modes.md)**  
  Common symptoms, likely causes, and safe fixes (with links to runbooks).
- **[Service not ready (runbook)](runbooks/service-not-ready.md)**  
  Triage steps when the API fails readiness.
- **[Empty recs (runbook)](runbooks/empty-recs.md)**  
  Common causes and safe remediations.
- **[Pipelines runbooks](../recsys-pipelines/docs/operations/runbooks/pipeline-failed.md)**  
  Day-2 operations for the offline layer.

</div>

## Read next

- Production readiness checklist: [`operations/production-readiness-checklist.md`](production-readiness-checklist.md)
- Pipelines SLOs and freshness: [`recsys-pipelines/docs/operations/slos-and-freshness.md`](../recsys-pipelines/docs/operations/slos-and-freshness.md)
