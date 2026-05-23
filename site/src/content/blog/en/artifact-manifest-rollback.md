---
title: "Artifact manifest rollback for recommendation systems"
description: "How artifact manifest rollback keeps recommendation models, features, and serving artifacts recoverable without turning every issue into a service rollback."
language: "en"
pubDate: "2026-05-09"
translationKey: "artifact-manifest-rollback"
tags: ["artifact manifest", "rollback", "pipelines"]
---

Artifact manifest rollback is the difference between recovering a recommendation system and redeploying everything
because one model, feature table, or rules bundle was wrong. In a serving system with frequent ranking and data changes,
the active artifact version needs to be explicit.

The service binary is only one part of the recommendation path. The response can also depend on model files, feature
snapshots, candidate indexes, config rules, and fallback definitions. If those artifacts move without a clear manifest,
operators lose the ability to explain or reverse production behavior.

## What an artifact manifest should identify

An artifact manifest is a pointer to the versioned inputs the serving layer is allowed to use. It should be small enough
to review and specific enough to reproduce a served response.

At minimum, the manifest should identify:

- model or scoring artifact versions
- candidate or index versions
- feature snapshot or feature extraction versions
- rule and config versions
- build or pipeline run metadata
- freshness timestamps
- validation status

The manifest does not have to contain every artifact payload. It should point to the exact versions that make up a
serving release.

## Why service rollback is not enough

Rolling back the API deployment can be the wrong lever if the failure came from data or artifacts. A previous binary may
still load the same bad manifest. A new binary may be healthy while the active index is stale.

Artifact manifest rollback gives the operator a narrower recovery path. Instead of redeploying the service, the team can
restore the last known-good manifest pointer, confirm the service reads it, and watch guardrails return to the expected
range.

This also improves incident review. The team can separate code defects from pipeline defects, stale data, and bad
ranking configuration.

## Artifact manifest rollback flow

A practical rollback flow is short and explicit:

1. Identify the failing rollout, affected tenant or surface, and active manifest version.
2. Confirm the guardrail or evaluation result that triggered rollback.
3. Select the previous known-good manifest from the decision record.
4. Move the active manifest pointer back to that version.
5. Confirm readiness, freshness, response shape, and exposure logging.
6. Record the restored version and the reason for rollback.

The technical docs cover this workflow in the
[artifacts and pipelines guide](/documentation/technical/artifacts-and-pipelines/) and the
[stale artifact manifest runbook](/documentation/technical/operations/runbooks/stale-artifact-manifest/).

## Backfills need the same discipline

Backfills are useful, but they can also make production state harder to reason about. If a backfill updates features,
indexes, or artifacts, it should create a new manifest candidate instead of silently changing the active serving inputs.

That gives evaluators and operators a clean decision point: validate the new manifest, ship it, hold it, or roll it back.

## Connect rollback to evaluation

Rollback should be part of the evaluation plan, not an emergency invention. The team should know which guardrails matter,
which manifest version is current, and which previous version is safe before traffic moves.

For the product-level workflow, see the [evaluation page](/evaluation/). For operational recovery, use the
[rollback config and rules runbook](/documentation/technical/operations/runbooks/rollback-config-rules/).
