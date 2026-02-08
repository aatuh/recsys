---
diataxis: how-to
tags:
  - how-to
  - developer
  - recsys-engineering
  - pipelines
---
# How-to: add a new signal end-to-end
This guide shows how to how-to: add a new signal end-to-end in a reliable, repeatable way.


## Who this is for

- Engineers extending RecSys with a new signal that affects ranking
- Teams moving from “we can run it” to “we can improve it safely”

## What you will get

- A concrete, end-to-end checklist: `port` → `pipeline artifact` → `manifest` → `serving` → `eval`
- The exact places in the codebase where new signals are wired today
- Verification points and safe rollback options

!!! info "Scope"
    This guide focuses on **artifact/manifest mode**, because it is the safest way to ship/rollback signal changes.
    You can implement signals in DB-only mode too, but the “publish pointer last” safety story is weaker.

## Prereqs

- You can run a production-like setup locally: [production-like run (pipelines → object store → ship/rollback)](../tutorials/production-like-run.md)
- You understand the current ranking pipeline order: [Ranking & constraints reference](../recsys-algo/ranking-reference.md)
- You have the minimum instrumentation in place for evaluation: [Minimum instrumentation spec (for credible evaluation)](../reference/minimum-instrumentation.md)

## Step 0: classify your signal (so you change the right layer)

Most signals fit one of these shapes:

- **Candidate retrieval signal**: changes which items are considered.
  - Example: “category trending” adds candidates to the pool.
- **Scoring signal**: changes how candidates are ranked.
  - Example: “user-to-item affinity” contributes a score term.
- **Post-ranking constraint**: changes ordering or membership after scoring.
  - Example: brand/category caps, diversity (MMR), blocklists.

In `recsys-algo`, scoring currently has:

- a baseline term (`pop_raw` from the primary candidate source)
- optional co-visitation
- a similarity bucket (`embedding` / `collaborative` / `content` / `session`)
- an optional personalization multiplier

See the formal spec: [Scoring model specification (recsys-algo)](../recsys-algo/scoring-model.md)

## Step 1: define (or reuse) the store port in `recsys-algo`

If your signal is already represented by an existing port, reuse it. Otherwise:

1. Add a new interface to `recsys-algo/model/model.go`.
1. Extend the `algorithm` layer to call it:
   - candidate retrieval: `recsys-algo/algorithm/candidate_sources.go`
   - scoring: `recsys-algo/algorithm/scoring.go` (and possibly `selectSimilarity`)
   - request-time gating/weights: `recsys-algo/algorithm/signals.go`
1. Add unit tests in `recsys-algo/algorithm/*_test.go` to lock determinism (stable ordering on ties).

Tip: when a capability is not available, return `model.ErrFeatureUnavailable` and keep serving using other signals.

See: [Store ports](../recsys-algo/store-ports.md)

## Step 2: add an artifact type + compute pipeline (recsys-pipelines)

If the signal is computed offline, add it as a new artifact type.

Follow the module guide: [How-to: Add a new artifact type](../recsys-pipelines/docs/how-to/add-artifact-type.md)

In practice, you will touch (at least) these areas:

- Domain type and model:
  - `recsys-pipelines/internal/domain/artifacts/types.go` (new `artifacts.Type`)
  - `recsys-pipelines/internal/domain/artifacts/models.go` (v1 struct + constructor)
- Compute:
  - `recsys-pipelines/internal/app/usecase/compute_<signal>.go`
- Validation:
  - `recsys-pipelines/internal/adapters/validator/builtin/validator.go`
- Publishing and manifest update:
  - `recsys-pipelines/internal/app/usecase/publish_artifacts.go`
  - `recsys-pipelines/internal/app/workflow/pipeline.go`

Non-negotiables:

- Deterministic output for the same canonical inputs
- Bounded runtime and memory use
- Manifest pointer updated last (two-phase publish)

Background: [Artifacts and versioning](../recsys-pipelines/docs/explanation/artifacts-and-versioning.md)

## Step 3: publish and verify the manifest contains your artifact

Run pipelines and confirm the manifest key is present.

1. Run a local pipelines job (example):

```bash
recsys-pipelines run --config pipelines.config.json
```

1. Verify the output layout and manifest pointer:

```bash
find /tmp/recsys-pipelines -maxdepth 5 -type f | head
```

The current manifest is the contract between pipelines and the service. See:

- Output layout: [Output layout (local filesystem)](../recsys-pipelines/docs/reference/output-layout.md)
- Manifest schema: [Manifest schema (JSON)](../reference/data-contracts/artifacts/manifest.schema.json)

## Step 4: make `recsys-service` consume the artifact

In artifact mode, the service reads `manifest.current` and loads known artifact types.

For a new artifact-backed signal, you will typically update:

- `api/internal/artifacts/types.go` (add the manifest key constant and the artifact struct + validator)
- `api/internal/store/artifact.go` (load the artifact URI and implement the relevant `recsys-algo/model` port)

Definition of Done for this step:

- Service starts even if your new artifact key is missing (treat as “signal unavailable”).
- When the artifact is present and valid, the port returns results deterministically.

## Step 5: expose a safe serving toggle (config)

Do not “silently” enable a new signal everywhere.

Pick one (or both):

- **Weight-gate**: introduce a new blend term or reuse an existing one and default it to `0`.
- **Mode-gate**: add a new algorithm mode (advanced; increases surface area).

Ensure the toggle is:

- documented in `reference/config/recsys-service.md`
- visible in traces/explain output (so you can prove what was active)

## Step 6: verify serving behavior

1. Start the service in artifact mode and point it at a manifest containing the new artifact.
2. Request recommendations with explain enabled (example):

```bash
curl -sS http://localhost:8000/v1/recommend \\
  -H 'Content-Type: application/json' \\
  -H 'X-RecSys-Tenant: demo' \\
  -d '{\"user_id\":\"u_1\",\"surface\":\"home\",\"k\":10,\"explain\":\"full\"}' | jq .
```

Verify:

- the response is non-empty
- the `explain` block includes your signal’s contribution (or a clear “unavailable” state)
- tie-breaking is stable (repeat the same request and compare ordering)

## Step 7: evaluate and ship safely

1. Generate an evaluation dataset (or use your pilot logs).
1. Run an offline comparison and guardrail checks:

```bash
recsys-eval run --dataset dataset.yml --config eval.yml
```

1. Use the report to decide ship/hold/rollback:

- Suite workflow: [How-to: run evaluation and make ship decisions](run-eval-and-ship.md)
- Decision playbook: [Decision playbook: ship / hold / rollback](../recsys-eval/docs/decision-playbook.md)

## Rollback options (use at least one)

- **Manifest rollback (artifact mode):** swap the manifest pointer back to a previous version.
- **Config rollback:** set the new signal weight back to `0` or disable the mode.
- **Rules rollback:** revert pinned/boost/block rules (if your change depended on rules).

See:

- Roll back the manifest safely: [How-to: Roll back to a previous artifact version](../recsys-pipelines/docs/how-to/rollback-manifest.md)
- Roll back artifacts safely: [How-to: Roll back artifacts safely](../recsys-pipelines/docs/how-to/rollback-safely.md)

## Common failure modes (and what to check first)

- **Serving says `SIGNAL_UNAVAILABLE`:** manifest key missing, wrong tenant/surface, artifact not published, or the
  store port is not wired.
- **Serving returns “empty recs”:** the new signal might have filtered candidates too aggressively or caps blocked all
  remaining items.
  - Runbook: [Runbook: Empty recs](../operations/runbooks/empty-recs.md)
- **Pipelines publish fails:** validator rejects artifact or version recompute mismatch; manifest should remain on the
  previous version (safe failure).
- **Non-deterministic ordering:** unstable iteration order over maps or missing tie-break sorting in a store backend.
  - Determinism notes: [Ranking & constraints reference](../recsys-algo/ranking-reference.md)

## Read next

- Store ports reference: [Store ports](../recsys-algo/store-ports.md)
- Pipelines artifact extension: [How-to: Add a new artifact type](../recsys-pipelines/docs/how-to/add-artifact-type.md)
- Artifacts and versioning: [Artifacts and versioning](../recsys-pipelines/docs/explanation/artifacts-and-versioning.md)
