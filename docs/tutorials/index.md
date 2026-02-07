---
tags:
  - tutorials
  - overview
  - developer
---

# Tutorials

## Who this is for

- People who want a guided, runnable walkthrough of RecSys (copy/paste steps).

## What you will get

- Local “first success” (hello recommendations)
- A full end-to-end loop (serve → log → evaluate → ship/rollback)
- A production-like run you can mirror in staging

--8<-- "_snippets/key-terms.list.snippet"
--8<-- "_snippets/key-terms.defs.one-up.snippet"

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

</div>

## Read next

- Quickstart (10 minutes): [`tutorials/quickstart.md`](quickstart.md)
- Local end-to-end tutorial: [`tutorials/local-end-to-end.md`](local-end-to-end.md)
- Integrate the serving API into your app: [`how-to/integrate-recsys-service.md`](../how-to/integrate-recsys-service.md)
