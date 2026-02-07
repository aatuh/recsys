---
tags:
  - overview
  - ml
  - evaluation
  - recsys-algo
  - recsys-eval
---

# RecSys engineering: start here

## Who this is for

- Recommendation engineers validating ranking behavior, attribution, and evaluation gates
- ML engineers who need to understand what is (and is not) in scope for the suite

## What you will get

- A curated reading order (10–30 minutes)
- The canonical “contracts” for ranking, logs, and evaluation
- The shortest path to reproduce behavior locally

--8<-- "_snippets/key-terms.list.snippet"
--8<-- "_snippets/key-terms.defs.one-up.snippet"

!!! info "RecEng terms"
    - **[Candidate](../project/glossary.md#candidate)**: an item considered for ranking.
    - **[Ranking](../project/glossary.md#ranking)**: ordering candidates into the returned list.

## Read this in order (recommended)

1. Suite mental model and data flow:
   - [How it works](../explanation/how-it-works.md)
2. Deterministic ranking behavior (what is guaranteed):
   - [Ranking reference](../recsys-algo/ranking-reference.md)
3. Data mode implications for evaluation and rollback:
   - [Choose your data mode](choose-data-mode.md)
   - [Data modes (details)](../explanation/data-modes.md)
4. Logging and attribution (what makes metrics trustworthy):
   - [Exposure logging & attribution](../explanation/exposure-logging-and-attribution.md)
   - [Data contracts (schemas)](../reference/data-contracts/index.md)
5. Default evaluation workflow and decision-making:
   - [Run eval & ship decisions](../how-to/run-eval-and-ship.md)
   - [Decision playbook](../recsys-eval/docs/decision-playbook.md)
6. Experimentation and iteration model:
   - [Experimentation model](../explanation/experimentation-model.md)
7. Scope boundaries:
   - [Capability matrix](../explanation/capability-matrix.md)
   - [Known limitations](known-limitations.md)

## Want runnable behavior fast?

- End-to-end loop (serve → log → eval): [`tutorials/local-end-to-end.md`](../tutorials/local-end-to-end.md)
- Production-like ship/rollback mechanics: [`tutorials/production-like-run.md`](../tutorials/production-like-run.md)

## Read next

- Add a signal end-to-end: [How-to: add a signal end-to-end](../how-to/add-signal-end-to-end.md)
- recsys-eval CI workflow:
  [Offline gate in CI](../recsys-eval/docs/workflows/offline-gate-in-ci.md)
- recsys-pipelines artifact rollback:
  [Roll back artifacts safely](../recsys-pipelines/docs/how-to/rollback-safely.md)
