---
diataxis: tutorial
tags:
  - tutorials
  - overview
  - developer
---
# Tutorials
In this tutorial you will follow a guided walkthrough and verify a working result.


## Who this is for

- People who want a guided, runnable walkthrough of RecSys (copy/paste steps).

## What you will get

- Local “first success” (hello recommendations)
- A full end-to-end loop (serve → log → evaluate → ship/rollback)
- A production-like run you can mirror in staging

## Choose a tutorial

<div class="grid cards" markdown>

- **[Quickstart (10 minutes)](quickstart.md)**  
  Fastest path to a non-empty `POST /v1/recommend` response and an exposure log.
- **[Local end-to-end](local-end-to-end.md)**  
  Full walkthrough: run the suite locally and produce a report.
- **[Production-like run](production-like-run.md)**  
  Practice the ship/hold/rollback workflow in a staging-like setup.
- **[Minimal pilot (DB-only)](minimal-pilot-db-only.md)**  
  A reduced setup for early integration and stakeholder demos.

- **[Verify determinism](verify-determinism.md)**  
  Prove stable ranking output for identical inputs.
- **[Verify joinability](verify-joinability.md)**  
  Prove exposures and outcomes can be joined for evaluation.

</div>

## Read next

- Quickstart (10 minutes): [Quickstart (10 minutes)](quickstart.md)
- Local end-to-end tutorial: [local end-to-end (service → logging → eval)](local-end-to-end.md)
- Integrate the serving API into your app: [How-to: integrate recsys-service into an application](../how-to/integrate-recsys-service.md)
