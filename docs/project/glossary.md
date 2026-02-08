---
diataxis: reference
tags:
  - project
  - reference
---
# Glossary

A small shared vocabulary used throughout the suite docs.

## Artifact

An immutable, version-addressed blob produced offline (for example: popularity, co-visitation, embeddings) and consumed
by `recsys-service`.

## Artifact/manifest mode

A deployment mode where `recsys-pipelines` publishes versioned artifacts and a mutable “current manifest pointer”, and
`recsys-service` reads the current manifest and fetches referenced blobs.

See: [Data modes: DB-only vs artifact/manifest](../explanation/data-modes.md)

## DB-only mode

A deployment mode where the service reads signals directly from Postgres tables (no offline artifact publish step).

DB-only mode is useful for early pilots because it minimizes moving parts, but it trades off offline reproducibility
and versioned “ship/rollback” via manifests.

See: [Data modes: DB-only vs artifact/manifest](../explanation/data-modes.md)

## Manifest

A small document that maps artifact types to artifact URIs for a `(tenant, surface)` pair. In artifact mode, the “current
manifest pointer” is what you ship and roll back.

## Freshness

An operational concept: “are recommendations based on recent-enough data?” In artifact/manifest mode, freshness is often
defined as “the current manifest was updated within an expected window”.

## Candidate

An item considered for ranking. Candidate generation prioritizes recall: “what items should we even consider?”

## Ranking

Scoring and ordering candidates to produce the final top-K list. Ranking prioritizes precision: “which of these are
best?”

## Exposure

An event that records what items were shown (and in what order) for a single recommendation request.

## Exposure log

The stream/file of exposure events written by `recsys-service` (typically JSONL). This is an input to `recsys-eval` and
one of the main audit artifacts for “what was shown”.

See: [Exposure logging and attribution](../explanation/exposure-logging-and-attribution.md)

## Outcome

An event that records what the user did after an exposure (click, conversion, etc.). Outcomes are attributed to exposures
by `request_id`.

## Outcome log

The stream of outcome events emitted by your product (clicks, conversions, etc.) and joined to exposures by
`request_id`.

See: [Exposure logging and attribution](../explanation/exposure-logging-and-attribution.md)

## Assignment

An event that records experiment bucket membership (experiment id + variant) used for A/B analysis.

## Request ID

A correlation identifier that ties together “recommend” responses, exposure logs, and outcome events. In this suite it is
the primary join key for evaluation.

## Segment

A cohort/slice label (default: `default`) used for rule scoping and evaluation breakdowns (guest vs returning, locale,
etc.).

## Control plane

The admin APIs and versioned documents (tenant config + rules) that control serving behavior without redeploying code.

## Ship

Promote a candidate change to “current” in a safe, reversible way (for example: update the manifest pointer, or update
tenant config/rules versions).

## Rollback

Revert “current” to a last-known-good version (config/rules rollback or manifest pointer rollback).

## Namespace

An application-defined bucket of recommendation logic/data. In this suite, `surface` typically acts as the namespace for
signals and rules.

## Surface

Where recommendations are shown (home, PDP, cart, …). Surface names should be stable; they scope signals, rules, and
evaluation slices.

## Tenant

An organization boundary for configuration and data isolation.

## recsys-svc

The Docker Compose container name used in this repo for the `recsys-service` API. In docs and architecture discussions,
prefer the module name `recsys-service`.


## Canonical events

Events stored in a normalized format that the rest of the pipeline relies on.

## Window

A time range that a job processes. In v1, windows are daily UTC buckets.

## Version

A deterministic identifier (SHA-256 hex) of an artifact payload excluding volatile build metadata.

## Checkpoint

A small state file that records the latest successfully processed window so incremental runs can skip work already done.

## Incremental run

A run mode that processes only new windows since the last checkpoint.

## Backfill

Re-processing a historical range of windows.

## Current manifest pointer

The mutable “what is live right now” location for a `(tenant, surface)` manifest.

## Registry

Storage for artifact records and current manifests.

## Object store

Storage for artifact blobs. In local mode, this is the filesystem.

## Idempotent

Safe to run multiple times without changing the result.


## Read next

- Project hub: [Project](index.md)
- Contributing: [Contributing](contributing.md)
